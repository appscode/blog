---
title: Read/Write Operations Splitting from Envoy Filter
date: "2024-12-13"
weight: 29
authors:
- Tauhedul Islam
tags:
- envoy
- postgresql-filters
- postgresql
- read_write-splitting
---

## Read/Write Operation Splitting 
In this blog, we'll discuss the read write operation splitting technique from Envoy PostgresQL filter.
\
In read-write splitting, we implement a mechanism where the read operations are sent to an upstream and the write operations are sent to a different upstream. This helps us balance the load on servers.
Here in Envoy, we implemented read-write splitting in the PostgreSQL proxy filter. Envoy by default maintains a downstream connection with the client and an upstream connection with the server.
To send the write operations to a different upstream we need to create an additional connection with an upstream server from the filter end, complete the initial handshake and authentication for this new connection, and reserve this connection. Then whenever we need, we may send data to this additional connection.

### Additional Connection
To create and handle this additional connection we implemented a common class `AdditionalConnection::UpstreamConnection` located in `envoy/additional_connection/additional_connection.h` file, 
which helps to create a new connection and communicate with that.\
To Use this class, we need to initialize it with the following object arguments.
1. ClusterManager
2. ClusterName
3. Dispatcher
4. AdditionalConnCallback

Here `ClusterManager` and `ClusterName` are used to choose the upstream cluster and host.\
Dispatcher is used to establish the connection with the chosen host.\
`AdditionalConnCallback` is used to get the response data from this additional connection.\
While working in the PostgreSQL filter, we inherited this `AdditionalConnCallback` class in our filter class and implemented its `writeCallback` function to get the response data in the PostgreSQL filter.
We also do the same for other filters.\
Then we may send the response from this filter to the downstream connection by using `read_callbacks_.connection().write()` method.\
\
Another important thing is getting the required objects for initializing the additional connection.
- ClusterManager: We found this from the `context` object found in the `config` part of the filter.
- ClusterName: We need to take this from the envoy configuration yaml. We've added a field in the proto filet to get this parameter. We'll see it in the sample yaml.
- Dispactcher: We may get it from the downstream connection object, which we can access by `read_callbacks_.connection()` from the filter.
- AdditionalConnCallback: This class need to be inherited an implemented in the filter class, and then we can use the filter object itself for this parameter.
- Username Passowrd: We also need a list of all usernames and passwords, this will be used for authentication of the additional connection.

### Sample Yaml
```yaml
- filters:
      - name: envoy.filters.network.postgres_proxy
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.postgres_proxy.v3alpha.PostgresProxy
          stat_prefix: egress_pg
          enable_sql_parsing: true
          terminate_ssl: true
          upstream_ssl: REQUIRE
          split_rw: true
          write_cluster_name: "pg_cluster_write"
          user_list:
            - username: "postgres"
              password: "X9DGwiB3awB;uA.X"
            - username: "pgpg"
              password: "pgpg"
      - name: envoy.filters.network.tcp_proxy
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          stat_prefix: pg_tcp
          cluster: pg_cluster_read
```

### Initial Handshake and Authentication
For the PostgreSQL filter, we implemented the functionalities to do the handshake with the additional connection maintaining the PostgreSQL protocol.
In PostgreSQL protocol:
- If the connection is SSL secured, the client sends an SSL Request packet in the beginning, and the server sends the response with `S` or `E`.
- Then the SSL handshake is done if the server sends `S`.
- The client sends the `Client Hello` packet, if the connection isn't SSL secured then this packet is sent as the first packet from the client.
- Then the Server sends the `Server Hello` packet, requesting the authentication packet from the client. It also mentions the authentication method and contains the `Salt` value for password hashing.
- The client sends the authentication packet containing the username and the password hash in the mentioned method with the `Salt` value.
- After authentication, the server sends an `Ok` packet and the connection is established.

Maintaining this, we reserve the `SSL Request` packet and the `Client Hello` packet when the default connection is getting established with the upstream. We also extract the username from the `Client Hello` packet and find its password from the given list of username-password from the Envoy configuration yaml.\
Then after the authentication is completed with the default upstream connection, we send the `SSL Request` packet or the `Client Hello` packet to the additional upstream depending on whether the upstream is SSL secured or not.\
If the upstream is SSL secured, then after receiving the `SSL Response` packet we call the `startSecureTransport` function of the additional connection and make it SSL secured. Then we send the `Client Hello` packet.\
After receiving the `Server Hello` packet, we generate the Authentication packet using the `username`, `password`, and the `Salt` value, and send this to additional connection. Note that, we've implemented MD5 based authentication method for now, other methods can be added in the samy way.\
After that receive the final response from the server and complete the connection process.\
In this handshake and authentication process with the additional connection, we don't send any response packet to downstream. As the downstream client has already completed these phase with the Envoy's default upstream server. We handle this process from the filter, send the packets from the filter and receive the packets from the filter.\
To implement this splitting from other filters, we also need to do the same.

### Query Handling
During the query phase, we determine if the query is of write type. If yes, we send the query request to the additional connection using `additional_connection.sendDataToUpstream()` method.
And return `FilterStatus::StopIteration` from the `onData` method of the PostgreSQL proxy filter, so that the data packed doesn't get send to the default upstream.\
Then we receive the response from the additional upstream in the `writeCallback` function mentioned above. We send this response to downstream by using `read_callbacks_.connection().write()` method.