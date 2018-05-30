import boto3,json,os,time
ec2Client = boto3.client('ec2')
autoscalingClient = boto3.client('autoscaling')
snsClient = boto3.client('sns')
lambdaClient = boto3.client('lambda')
s3Client = boto3.resource("s3")

def publishSNSMessage(snsMessage,snsTopicArn):
    response = snsClient.publish(TopicArn=snsTopicArn,Message=json.dumps(snsMessage),Subject='Rebalancing')

def checkEc2s(asgName):
    filters = [{  
    'Name': 'tag:aws:autoscaling:groupName',
    'Values': [asgName]
    }]
    ec2ContainerInstances = ec2Client.describe_instances(Filters=filters)
    print(str(ec2ContainerInstances))
    pendingEc2s = 0
    activeEc2s = 0
    for i in range(len(ec2ContainerInstances['Reservations'])):
        instance = ec2ContainerInstances['Reservations'][i]['Instances'][0]
        print(str(instance['State']['Name']))
        print(str(instance))
        if instance['State']['Name'] == 'disabling':
            pendingEc2s = pendingEc2s + 1
        elif instance['State']['Name'] == 'pending':
            pendingEc2s = pendingEc2s + 1
        elif instance['State']['Name'] == 'running':
            activeEc2s = activeEc2s + 1
    print("Active EC2s: ",activeEc2s)
    return pendingEc2s

def generateRKEConfig(asgName, instanceUser, keyName, FQDN):
    filters = [{  
    'Name': 'tag:aws:autoscaling:groupName',
    'Values': [asgName]
    }]
    ec2ContainerInstances = ec2Client.describe_instances(Filters=filters)

    rkeConfig = (' default k8s version: v1.8.9-rancher1-1.\n'
                ' # default network plugin: flannel.\n'
                ' ignore_docker_version: true.\n'
                ' .\n'
                ' nodes:.\n'
                ' .\n')

    for i in range(len(ec2ContainerInstances['Reservations'])):
        instance = ec2ContainerInstances['Reservations'][i]['Instances'][0]
        if instance['State']['Name'] == 'running':
            rkeConfig += ('    - address: ' + instance['PublicIpAddress'] + '.\n'
                                '  user: ' + instanceUser + '.\n'
                                '  role: [controlplane,etcd,worker].\n'
                                '  ssh_key_path: ' + keyName + '.\n'
                            ' .\n')

            rkeConfig += (' addons: |-.\n'
            '---.\n'
            'kind: Namespace.\n'
            'apiVersion: v1.\n'
            'metadata:.\n'
            '    name: cattle-system.\n'
            '---.\n'
            'kind: ServiceAccount.\n'
            'apiVersion: v1.\n'
            'metadata:.\n'
            '    name: cattle-admin.\n'
            '    namespace: cattle-system.\n'
            '---.\n'
            'kind: ClusterRoleBinding.\n'
            'apiVersion: rbac.authorization.k8s.io/v1.\n'
            'metadata:.\n'
            '    name: cattle-crb.\n'
            '    namespace: cattle-system.\n'
            'subjects:.\n'
            '- kind: ServiceAccount.\n'
            '    name: cattle-admin.\n'
            '    namespace: cattle-system.\n'
            'roleRef:.\n'
            '    kind: ClusterRole.\n'
            '    name: cluster-admin.\n'
            '    apiGroup: rbac.authorization.k8s.io.\n'
            '---.\n'
            'apiVersion: v1.\n'
            'kind: Service.\n'
            'metadata:.\n'
            '    namespace: cattle-system.\n'
            '    name: cattle-service.\n'
            '    labels:.\n'
            '    app: cattle.\n'
            'spec:.\n'
            '    ports:.\n'
            '    - port: 80.\n'
            '    targetPort: 80.\n'
            '    protocol: TCP.\n'
            '    name: http.\n'
            '    - port: 443.\n'
            '    targetPort: 443.\n'
            '    protocol: TCP.\n'
            '    name: https.\n'
            '    selector:.\n'
            '    app: cattle.\n'
            '---.\n'
            'apiVersion: extensions/v1beta1.\n'
            'kind: Ingress.\n'
            'metadata:.\n'
            '    namespace: cattle-system.\n'
            '    name: cattle-ingress-http.\n'
            '    annotations:.\n'
            '    nginx.ingress.kubernetes.io/proxy-connect-timeout: 30.\n'
            '    nginx.ingress.kubernetes.io/proxy-read-timeout: 1800.\n'
            '    nginx.ingress.kubernetes.io/proxy-send-timeout: 1800.\n'
            'spec:.\n'
            '    rules:.\n'
            '    - host: ' + FQDN + '.\n'
            '    http:.\n'
            '        paths:.\n'
            '        - backend:.\n'
            '            serviceName: cattle-service.\n'
            '            servicePort: 80.\n'
            '    tls:.\n'
            '    - secretName: cattle-keys-ingress.\n'
            '    hosts:.\n'
            '    - ' + FQDN + '.\n'
            '---.\n'
            'kind: Deployment.\n'
            'apiVersion: extensions/v1beta1.\n'
            'metadata:.\n'
            '    namespace: cattle-system.\n'
            '    name: cattle.\n'
            'spec:.\n'
            '    replicas: 1.\n'
            '    template:.\n'
            '    metadata:.\n'
            '        labels:.\n'
            '        app: cattle.\n'
            '    spec:.\n'
            '        serviceAccountName: cattle-admin.\n'
            '        containers:.\n'
            '        - image: rancher/rancher:latest.\n'
            '        imagePullPolicy: Always.\n'
            '        name: cattle-server.\n'
            '        ports:.\n'
            '        - containerPort: 80.\n'
            '            protocol: TCP.\n'
            '        - containerPort: 443.\n'
            '            protocol: TCP.\n'
            '        volumeMounts:.\n'
            '        - mountPath: /etc/rancher/ssl.\n'
            '            name: cattle-keys-volume.\n'
            '            readOnly: true.\n'
            '        volumes:.\n'
            '        - name: cattle-keys-volume.\n'
            '        secret:.\n'
            '            defaultMode: 420.\n'
            '            secretName: cattle-keys-server.\n'
            ' .\n')
    return rkeConfig

def lambda_handler(event, context):
    instanceUser=os.environ['InstanceUser']
    keyName=os.environ['KeyName']
    FQDN=os.environ['FQDN']
    rkeS3Bucket=os.environ['rkeS3Bucket']

    snsTopicArn=event['Records'][0]['Sns']['TopicArn']
    snsMessage=json.loads(event['Records'][0]['Sns']['Message'])

    lifecycleHookName=snsMessage['LifecycleHookName']
    lifecycleActionToken=snsMessage['LifecycleActionToken']
    asgName=snsMessage['AutoScalingGroupName']

    pendingEc2s=checkEc2s(asgName);
    if pendingEc2s==0:
        rkeConfig = generateRKEConfig(asgName,instanceUser,keyName)
        print("RKE Config:")
        print(rkeConfig)
        try:
            encoded_payload = rkeConfig.encode("utf-8")
            s3Client.Bucket(rkeS3Bucket).put_object(Key=config.yaml, Body=encoded_payload)
            response = autoscalingClient.complete_lifecycle_action(LifecycleHookName=lifecycleHookName,AutoScalingGroupName=asgName,LifecycleActionToken=lifecycleActionToken,LifecycleActionResult='CONTINUE')
        except BaseException as e:
            print(str(e))
    elif pendingEc2s>=1:
        time.sleep(5)
        publishSNSMessage(snsMessage,snsTopicArn)