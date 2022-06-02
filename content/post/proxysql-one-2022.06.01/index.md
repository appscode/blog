---
title: Load Balance MySQL Group Replication with TLS secured ProxySQL Cluster
date: 2022-06-02
weight: 20
authors:
- Tasdidur Rahman
tags:
- Cloud-native platform
- Kubernetes
- Database
- Kubernetes ProxySQL
- Run Production-grade Database
- MySQL
- ProxySQL
- KubeDB
---

#### Overview

ProxySQL is an open source high performance, high availability, database protocol aware proxy for MySQL. To know more about ProxySQL, you may refer to the [Link](https://kubedb.com/docs/v2022.05.24/guides/proxysql/overview/overview/).

From the KubeDB release `v2022.05.24` we have added ProxySQL support for KubeDB MySQL. Now you can provision a ProxySQL server or cluster of ProxySQL servers with declarative yamls using KubeDB operator.

With KubeDB operator, you can provision ProxySQL with much less effort than usual. To establish connection between KubeDB ProxySQL and KubeDB MySQL all you need to do is just mention the reference of MySQL object in your ProxySQL yaml. Which means, <br>
* You don't need to create monitor user in the MySQL server on your own
* You don't need to configure `mysql_servers` table manually in your ProxySQL
* You don't need to run `addition_to_sys.sql` in the MySQL server yourself
* If you are willing to deploy a ProxySQL cluster, you also don't need to configure `proxysql_servers` table manually
* And also if you want to bootstrap your ProxySQL servers with your own custom configuration, you just have to put that in a secret and mention that in the ProxySQL yaml, nothing else!

In this blog post I will try to show the basic setup and will test some functionalities. We will follow these steps below,
1) Install KubeDB
2) Deploy MySQL Group Replication
3) Deploy ProxySQL cluster
4) Send load over ProxySQL cluster and observe load distribution
5) Observe SSL status over ProxySQL frontend and backend connection


#### Install KubeDB

We need to have KubeDB operator version `v2022.05.24` or later to test ProxySQL. To install KubeDB in your cluster, you may follow this [Link](https://kubedb.com/docs/v2022.05.24/setup/)

#### Install cert manager

If you don't want TLS secured connection for your ProxySQL or MySQL you can skip this part. As we are going to test TLS secured connections, we need to install cert manager in our cluster first. You can install cert manager operator in your cluster with this following command, <br>
```shell
$ kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.2/cert-manager.yaml
```

#### Create issuer

We also need an issuer to generate certs and keys for TLS secured connections.
* Start off by generating our ca-certificates using openssl,
```shell
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mysql/O=kubedb"
```
* Create a secret using the certificate files we have just generated,
```shell
kubectl create secret tls my-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```
* Create an Issuer using the `my-ca` secret that holds the ca-certificate we have just created,
```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: my-issuer
  namespace: demo
spec:
  ca:
    secretName: my-ca
```
* Now apply the issuer yaml,
```shell
$ kubectl apply -f issuer.yaml
issuer.cert-manager.io/my-issuer created
```
#### Setup KubeDB MySQL

Now we need KubeDB MySQL servers in our cluster on which we would test our ProxySQL functionalities. To know details about KubeDB MySQL, you may refer to this [Link](https://kubedb.com/docs/v2022.05.24/guides/mysql/) <br>

* Let the name of our MySQL object be `mysql-server` in the namespace  `demo`. We are choosing `5.7.36` as the `MySQL Version`, and would like to create a `MySQL Group Replication` with `3` nodes. We want to secure our MySQL server connections with TLS as well. By applying  the following yaml we can get our target KubeDB MySQL.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "5.7.36"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: my-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
  terminationPolicy: WipeOut
```

* Now apply the yaml and wait for mysql to be ready.
```shell
$ kubectl apply -f mysql.yaml
mysql.kubedb.com/mysql-server created

$ kubectl get pods -n demo | grep mysql-server
mysql-server-0            2/2     Running   0          6m
mysql-server-1            2/2     Running   0          5m
mysql-server-2            2/2     Running   0          5m

$ kubectl get mysql -n demo 
NAME           VERSION   STATUS   AGE
mysql-server   5.7.36    Ready    6m
```
Our MySQL is ready now.

#### Deploy ProxySQL
Let's create a ProxySQL cluster with name `proxy-mysql-server` in the `demo` namespace. As we are currently supporting version `2.3.2` only, so set the version accordingly. In the `spec.backend` section we need to mention necessary information about our MySQL server. We just need to mention the MySQL object reference and replica count. We can skip the `spec.tls` section if we don't want to secure the proxysql frontend connections, but here we have mentioned that as we are going to test the functionality. It's an enterprise feature, so you need to have the KubeDB enterprise operator in your cluster.

* Let's create the yaml first.
```yaml 
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-mysql-server
  namespace: demo
spec:
  version: "2.3.2"
  replicas: 3 
  mode: GroupReplication
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 256Mi
    storageClassName: standard
  storageType: Durable
  backend:
    ref:
      apiGroup: "kubedb.com"
      kind: MySQL
      name: mysql-server
    replicas: 3
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: my-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  terminationPolicy: WipeOut
```
* Now apply the yaml and wait for ProxySQL to be ready.
```shell
$ kubectl apply -f proxysql.yaml
proxysql.kubedb.com/proxy-mysql-server created

$ kubectl get pods -n demo | grep proxy
proxy-mysql-server-0            1/1     Running   0          3m
proxy-mysql-server-1            1/1     Running   0          3m
proxy-mysql-server-2            1/1     Running   0          3m

$ kubectl get proxysql -n demo                                                                                                         
NAME                 VERSION   STATUS   AGE
proxy-mysql-server   2.3.2     Ready    4m

```

We are now ready to test our ProxySQL functionalities. We will follow the below steps to check our functionalities, <br>
* [Create a test user, test database and table in MySQL server](#create-test-elements)
* [Put an entry for the test user in ProxySQL server](#user-entry)
* [Send load over the ProxySQL cluster as the test user](#send-load)
* [Check the load distribution over the ProxySQL cluster](#check-load-distribution)
* [Check TLS on the ProxySQL backend and frontend connection](#check-tls)


#### <a id="create-test-elements"></a>Create test user, test database and table in MySQL server
* Exec into the MySQL primary pod with the following command, we will end up with something like this,
```shell
~ $ kubectl exec -it -n demo mysql-server-0 -- bash                            
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
root@mysql-server-0:/#
```
* Log in to the MySQL console with root password.
```shell
root@mysql-server-0:/# mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 639
Server version: 5.7.36-log MySQL Community Server (GPL)

Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql>
```
* Let's get the current users.
```shell
mysql> select user from mysql.user;
+---------------+
| user          |
+---------------+
| proxysql      |
| repl          |
| root          |
| mysql.session |
| mysql.sys     |
| root          |
+---------------+
7 rows in set (0.00 sec)
```
> Note : The user `proxysql` is the monitor user for ProxySQL created by KubeDB operator
* Now create the test user, the database and the table and grant privileges to the test user.
```shell
mysql> create user 'test'@'%' identified by 'test' REQUIRE SSL;
Query OK, 0 rows affected (0.00 sec)

mysql> select user from mysql.user;
+---------------+
| user          |
+---------------+
| proxysql      |
| repl          |
| root          |
| test          |
| mysql.session |
| mysql.sys     |
| root          |
+---------------+
7 rows in set (0.00 sec)

mysql> create database random;
Query OK, 1 row affected (0.00 sec)

mysql> use random;
Database changed

mysql> create table randomtb(color varchar(120), price integer, primary key(color));
Query OK, 0 rows affected (0.02 sec)

mysql> grant all privileges on random.* to 'test'@'%';
Query OK, 0 rows affected (0.00 sec)

mysql> flush privileges;
Query OK, 0 rows affected (0.01 sec)
```

#### <a id="user-entry"></a> Put entry for test user in ProxySQL

* Exec into any ProxySQL pod and log into the proxysql admin console,
```shell
~ $ kubectl exec -it -n demo proxy-mysql-server-0 -- bash                  
root@proxy-mysql-server-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 3378
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]>
```

* Run the following command and put entry for the test user.
```shell 
MySQL [(none)]> replace into mysql_users(username,password,active,use_ssl,default_hostgroup, max_connections) values('test','test',1,1,2,200);
Query OK, 1 row affected (0.003 sec)

MySQL [(none)]> SAVE MYSQL USERS TO DISK;
Query OK, 0 rows affected (0.013 sec)

MySQL [(none)]> LOAD MYSQL USERS TO RUNTIME;
Query OK, 0 rows affected (0.002 sec)
```

* Let's check the current users.
```shell
MySQL [(none)]> select username, active, use_ssl from runtime_mysql_users;
+----------+--------+---------+
| username | active | use_ssl |
+----------+--------+---------+
| proxysql | 1      | 0       |
| root     | 1      | 0       |
| test     | 1      | 1       |
| root     | 1      | 0       |
| proxysql | 1      | 0       |
| test     | 1      | 1       |
+----------+--------+---------+
6 rows in set (0.004 sec)
```

* Also let's check the current stats for ProxySQL servers.
```shell
MySQL [(none)]> select hostname, Queries, Client_Connections_created from stats_proxysql_servers_metrics; 
+---------------------------------------------------+---------+----------------------------+
| hostname                                          | Queries | Client_Connections_created |
+---------------------------------------------------+---------+----------------------------+
| proxy-mysql-server-2.proxy-mysql-server-pods.demo | 0       | 0                          |
| proxy-mysql-server-1.proxy-mysql-server-pods.demo | 0       | 0                          |
| proxy-mysql-server-0.proxy-mysql-server-pods.demo | 0       | 0                          |
+---------------------------------------------------+---------+----------------------------+
3 rows in set (0.001 sec)
```

We are now ready to send loads to the ProxySQL servers with test user.

#### <a id="send-load"></a> Send load over ProxySQL cluster
To send loads to ProxySQL servers, first we need to create a pod from which we will connect to the ProxySQL service and run script from that pod.
* Let's create a pod containing basic ubuntu image with the following yaml,
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: ubuntu
  name: ubuntu
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ubuntu
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: ubuntu
    spec:
      containers:
      - image: ubuntu
        name: ubuntu
        command: ["/bin/sleep", "3650d"]
        resources: {}
```
* Deploy this to the cluster.
```shell
$ kubectl apply -f deployment.yaml
deployment.apps/ubuntu created

$ kubectl get pods -n demo | grep ubuntu
ubuntu-6c6d9795f4-sjn8p   1/1     Running   0          1m
```

* Now exec into the pod, install necessaries, create the bash script for sending load and run the script.
```shell
~ $ kubectl exec -it -n demo ubuntu-6c6d9795f4-sjn8p -- bash 

root@ubuntu-6c6d9795f4-sjn8p:/# apt-get update
... ... ...
... ... ...

root@ubuntu-6c6d9795f4-sjn8p:/# apt-get install -y mysql-client
... ... ...
... ... ... 

root@ubuntu-6c6d9795f4-sjn8p:/# apt-get install nano
... ... ...
... ... ...

root@ubuntu-6c6d9795f4-sjn8p:/# nano script.sh

#paste the follwing script in the pop-up window
COUNTER=0

USER='test'
PROXYSQL_NAME='proxy-mysql-server'
NAMESPACE='demo'
PASS='test'

VAR="x"

while [  $COUNTER -lt 100 ]; do
    let COUNTER=COUNTER+1
    VAR=a$VAR
    mysql -u$USER -h$PROXYSQL_NAME.$NAMESPACE.svc -P6033 -p$PASS -e 'select 1;' > /dev/null
    mysql -u$USER -h$PROXYSQL_NAME.$NAMESPACE.svc -P6033 -p$PASS -e "INSERT INTO random.randomtb(color, price) VALUES ('$VAR',50);" > /dev/null
    mysql -u$USER -h$PROXYSQL_NAME.$NAMESPACE.svc -P6033 -p$PASS -e "select * from random.randomtb;" > /dev/null
    sleep 0.0001
done
#script ends

root@ubuntu-6c6d9795f4-sjn8p:/# chmod +x script.sh
root@ubuntu-6c6d9795f4-sjn8p:/# ./script.sh
```
We have successfully sent the loads, now it's time to check what happened in the ProxySQL end.

#### <a id="check-load-distribution"></a> Check load distribution
Let's exec into any of the ProxySQL pods and log in to the admin console as before.<br>
* Now run the following command,
```shell
MySQL [(none)]> select hostname, Queries, Client_Connections_created from stats_proxysql_servers_metrics; 
+---------------------------------------------------+---------+----------------------------+
| hostname                                          | Queries | Client_Connections_created |
+---------------------------------------------------+---------+----------------------------+
| proxy-mysql-server-2.proxy-mysql-server-pods.demo | 194     | 97                         |
| proxy-mysql-server-1.proxy-mysql-server-pods.demo | 208     | 104                        |
| proxy-mysql-server-0.proxy-mysql-server-pods.demo | 198     | 99                         |
+---------------------------------------------------+---------+----------------------------+
3 rows in set (0.001 sec)
```
We can clearly see that the queries have been properly distributed over the ProxySQL cluster. If you can not see the load distribution in the first place, wait a bit for the cluster sync up and then hit the query again.

* Let's check how it was distributed to the MySQL servers.
```shell
MySQL [(none)]> select hostgroup, srv_host, Queries from stats_mysql_connection_pool;
+-----------+---------------------------------------+---------+
| hostgroup | srv_host                              | Queries |
+-----------+---------------------------------------+---------+
| 2         | mysql-server-0.mysql-server-pods.demo | 32      |
| 3         | mysql-server-0.mysql-server-pods.demo | 21      |
| 3         | mysql-server-1.mysql-server-pods.demo | 24      |
| 3         | mysql-server-2.mysql-server-pods.demo | 25      |
+-----------+---------------------------------------+---------+
4 rows in set (0.005 sec)
```
We can see that queries were properly distributed over the MySQL servers as well.

#### <a id="check-tls"></a> Check TLS

* Exec into the ubuntu pod, open in MySQL console with test user credential and connect to ProxySQL service.
```shell
$ kubectl exec -it -n demo ubuntu-6c6d9795f4-sjn8p -- bash 
root@ubuntu-6c6d9795f4-sjn8p:/# mysql -utest -ptest -hproxy-mysql-server.demo.svc -P6033
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 1468
Server version: 8.0.27 (ProxySQL)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> 
``` 

* Let's check the backend ssl with the following command,
```shell
mysql> show session status like 'Ssl_cipher';
+---------------+-----------------------------+
| Variable_name | Value                       |
+---------------+-----------------------------+
| Ssl_cipher    | ECDHE-RSA-AES256-GCM-SHA384 |
+---------------+-----------------------------+
1 row in set (0.00 sec)
```
We can see that the backend connection is TLS secured

* Let's check the frontend ssl with the following command,
 ```shell
 mysql> status;
--------------
mysql  Ver 8.0.29-0ubuntu0.22.04.2 for Linux on x86_64 ((Ubuntu))

Connection id:		1468
Current database:	information_schema
Current user:		test@10.244.0.21
SSL:			Cipher in use is TLS_AES_256_GCM_SHA384
Current pager:		stdout
Using outfile:		''
Using delimiter:	;
Server version:		8.0.27 (ProxySQL)
Protocol version:	10
Connection:		proxy-mysql-server.demo.svc via TCP/IP
Server characterset:	latin1
Db     characterset:	utf8
Client characterset:	latin1
Conn.  characterset:	latin1
TCP port:		6033
Binary data as:		Hexadecimal
Uptime:			2 hours 47 min 21 sec

Threads: 1  Questions: 202  Slow queries: 202
--------------
```

We can see the ssl status : `SSL:			Cipher in use is TLS_AES_256_GCM_SHA384`. Which means our frontend connection is also TLS secured.

#### Conclusion
In this blog post we have set up the basic connection and tested some basic functionalities of KubeDB ProxySQL. In the upcoming blog post we will try to cover the custom configuration and failover recovery of ProxySQL servers in ProxySQL cluster.


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
