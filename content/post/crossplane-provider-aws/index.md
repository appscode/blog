---
title: Deploy AWS Databases with KubeDB Provider-AWS
date: "2023-12-20"
weight: 26
authors:
- SK Ali Arman
tags:
- crossplane
- aws
- kubedb
- kubernetes
---

 # Crossplane

KubeDB is now a [Crossplane](https://www.crossplane.io/) distribution for Hyper Clouds. Crossplane connects your Kubernetes cluster to external, non-Kubernetes resources, and allows platform teams to build custom Kubernetes APIs to consume those resources. We have introduced providers for AWS.

You need [crossplane](https://docs.crossplane.io/v1.14/) already installed in your cluster. This will allow KubeDB users to provision and manage Cloud provider managed databases in a Kubernetes native way.


 ## Install Crossplane 

Add crossplane helm repository

```bash
helm repo add crossplane https://charts.crossplane.io/stable
helm repo update
```

Install crossplane

```bash
helm upgrade -i crossplane crossplane/crossplane -n crossplane-system --create-namespace
```

Check the installation with the following command. You will see two pod with prefix name crossplane in case of successfull installation. 

```bash
kubectl get pod -n crossplane-system
```

 ## Install KubeDB AWS Provider

Install the AWS provider into Kubernetes cluster with helm chart.

```bash
helm upgrade -i kubedb-provider-aws \
  oci://ghcr.io/appscode-charts/kubedb-provider-aws \
  --version=v2023.12.11 \
  -n crossplane-system --create-namespace
```


 ### Setup Provider Config

Create AWS access key and secret key from AWS IAM. You can see [AWS documentaion](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html).

Create a text file containing the AWS account aws_access_key_id and aws_secret_access_key with the following command.

```bash
export AWS_ACCESS_KEY_ID=<your_access_key>
export AWS_SECRET_ACCESS_KEY=<your_secret_access_key>
echo "
[default]
aws_access_key_id = $AWS_ACCESS_KEY_ID
aws_secret_access_key = $AWS_SECRET_ACCESS_KEY
" > aws-credentials.txt
```

Create a Kubernetes secret with the AWS credentials.

```bash
kubectl create secret  generic aws-secret -n crossplane-system --from-file=creds=./aws-credentials.txt
```

Create the ProviderConfig with the following yaml file

```bash
cat <<EOF | kubectl apply -f -
apiVersion: aws.kubedb.com/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: aws-secret
      key: creds
      namespace: crossplane-system
EOF
```

Check provider-config with the following command.

```bash
kubectl get providerconfig.aws.kubedb.com
```

 ### Create Postgres Instance

Create VPC

```bash
cat <<EOF | kubectl apply -f -
apiVersion: ec2.aws.kubedb.com/v1alpha1
kind: VPC
metadata:
  name: example
spec:
  forProvider:
    region: us-east-2
    cidrBlock: 172.16.0.0/16
    tags:
      Name: DemoVpc
EOF
```

Create three subnets

```bash
cat <<EOF | kubectl apply -f -
apiVersion: ec2.aws.kubedb.com/v1alpha1
kind: Subnet
metadata:
  name: example-subnet1
spec:
  forProvider:
    region: us-east-2
    availabilityZone: us-east-2b
    vpcIdRef:
      name: example
    cidrBlock: 172.16.10.0/24
EOF
```

```bash
cat <<EOF | kubectl apply -f -
apiVersion: ec2.aws.kubedb.com/v1alpha1
kind: Subnet
metadata:
  name: example-subnet2
spec:
  forProvider:
    region: us-east-2
    availabilityZone: us-east-2c
    vpcIdRef:
      name: example
    cidrBlock: 172.16.20.0/24
EOF
```

```bash
cat <<EOF | kubectl apply -f -
apiVersion: ec2.aws.kubedb.com/v1alpha1
kind: Subnet
metadata:
  name: example-subnet3
spec:
  forProvider:
    region: us-east-2
    availabilityZone: us-east-2a
    vpcIdRef:
      name: example
    cidrBlock: 172.16.30.0/24
EOF
```

Create Subnet Group

```bash
cat <<EOF | kubectl apply -f -
apiVersion: rds.aws.kubedb.com/v1alpha1
kind: SubnetGroup
metadata:
  name: example
spec:
  forProvider:
    region: us-east-2
    subnetIdRefs:
      - name: example-subnet1
      - name: example-subnet2
      - name: example-subnet3
    tags:
      Name: My DB subnet group
EOF
```

Create Secret

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: example-dbinstance
  namespace: crossplane-system
type: Opaque
data:
  password: MTIzNDU2Nzg=
EOF
```

Deploy the rds postgres instance

```bash
cat <<EOF | kubectl apply -f -
apiVersion: rds.aws.kubedb.com/v1alpha1
kind: Instance
metadata:
  annotations:
    meta.kubedb.com/example-id: rds/v1alpha1/instance
  labels:
    testing.kubedb.com/example-name: example-dbinstance
  name: example
spec:
  forProvider:
    region: us-east-2
    engine: postgres
    engineVersion: "15.4"
    username: adminuser
    dbName: dbexample
    passwordSecretRef:
      key: password
      name: example-dbinstance
      namespace: crossplane-system
    instanceClass: db.t3.micro
    storageType: gp2
    allocatedStorage: 20
    dbSubnetGroupName: example
    backupRetentionPeriod: 0
    backupWindow: "09:46-10:16"
    maintenanceWindow: "Mon:00:00-Mon:03:00"
    publiclyAccessible: false
    skipFinalSnapshot: true
    storageEncrypted: true
    autoMinorVersionUpgrade: true
  writeConnectionSecretToRef:
    name: example-dbinstance-out
    namespace: crossplane-system
EOF
```

 ### Provider AWS also supports

- DocumentDB
- Elasticache
- Dynamodb
- RDS
    - MariaDB
    - MySQL

 ## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).


