---
title: MySQL Filter in Envoy
date: "2024-12-13"
weight: 28
authors:
- Tauhedul Islam
tags:
- envoy
- mysql-filter
- mysql
- proxy
---

## MySQL Filter


The MySQL proxy filter decodes the wire protocol between the MySQL client and server. While using MySQL filter, all data
passes through this filter. We may decode, track and change communication packets from MySQL filter.
Using this advantage, we implemented some functionality in MySQL filter.
1. Terminate/Establish TLS conenction.
2. Logging the operations and export them.

In This blog, we'll write about process of handling TLS connection from MySQL filter.

### Sample Envoy Config

Here's an sample config yaml for a filter chain with MySQL proxy.

```yaml

filter_chains:
- filters:
  - name: envoy.filters.network.mysql_proxy
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.network.mysql_proxy.v3.MySQLProxy
      stat_prefix: mysql
  - name: envoy.filters.network.tcp_proxy
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
      stat_prefix: tcp
      cluster: "mysql_cluster"
```
We can see that, in the `filter chain` there are two filters, `mysql_filter` and `tcp_proxy`. MySQL server and
client uses a TCP connection to communicate with each other, that's why a `tcp_proxy` has been used here to establish a 
connection with the upstream server and transfer data. `tcp_proxy` should always be the last one in the filter chain.

### Basic Structure
MySQL filter is a network filter which inherits `Network::Filter` class. We need to implement several functions of this 
filter class in our MySQL filer. 
1. onData
2. onWrite
3. onNewConnection
4. addReadCallbacks

These are some primary functions we need to implement in any network filter class to have the basic functionalities of
the filter.

### How Filter Works
There's a filter manager class who is responsible for iterating over all the filters of the fi;lter chain when a new data
packet arrives from the upstream or downstream. 

- After the initialization of the filter, the `addReadCallbacks` function gets called to register a callback object in the filter, 
which we may use to call some functions of the parent class the `filterManager`
- When a new connection is established with the upstream, `onNewConnection` function gets called.
- When a data packet arrives from the downstream client, `onData` function gets called with the correspondig data packet.
- When a data packet arrives from the upstream servert, `onWrite` function gets called with the corresponding data packet.\

These are the basic functions a network filter should implement. Besides these, there are some other functions in the
MySQL filter to implement the decoding functionality and controlling tls connection. We'll discuss the TLS connection
controlling functionalities in this blog.

### TLS Management in MySQL Filter
To understanding the TLS connection management of MySQL proxy, we first need to understand the MySQL client server connection
protocol described [here](https://dev.mysql.com/doc/dev/mysql-server/8.4.3/page_protocol_connection_phase.html).

We have 2 fields available in the MySQL proxy filter to manage the TLS connection configuration.
1. terminate_downstream_tls
2. upstream_tls

The both fields are of `Bool` type.

If `terminate_downstream_tls` is true, that means there will be a TLS secured connection
from the client end, and we need to terminate that from the Envoy.

If the `upstream_tls` is true, that means we'll have to establish an TLS secured connection with the upstream from envoy.

Based on these, there might be 4 different configuration for TLS connection:
1. Downstream TLS + Upstream TLS
2. Downstream TLS + Upstream No-TLS
3. Downstream No-TLS + Upstream TLS
4. Downstream No-TLS + Upstream No-TLS


#### Downstream TLS + Upstream TLS:
In this configuration, we need to establish a TLS secured connection with the downstream client and another TLS secured 
connection with the upstream client. Envoy by default maintains 2 connections, one with the upstream and another with the
downstream. We just need to make them TLS secured by maintaining MySQL server/client protocol.
According to the MySQL protocol, the ssl handshake is done after a `ssl request` packet sent from the client.
Maintaining this protocol, when we receive this packet from the downstream client we call the `connection.startSecureTransport()`
function from the filter. By calling this function, Envoy will handle the `ssl exchange` part with the downstream client
using its own functionalities.

After Successfully establishing SSl connection with the downstream, we'll receive the `Handshake Response` packet from the
downstream. When we get this packet in the filter, before sending it to the upstream we need to establish the SSL secured 
connection with the upstream. The last packet sent to the upstream was the `ssl request` packet. So, the upstream server is
now waiting for the `SSL Exchange`. So we call `read_callbacks.startUpstreamSecureTransport()` function to complete the 
`ssl exchange` with the upstream server and make the upstream connection SSL secured. As stated earlier, this SSL key exchange
part will be handled by Envoy with its own functionalities, the packets won't come to the MySQL filter, and we don't need to worry
about them.

After the SSL handshake done in both end of Envoy, we can forward all other packet from downstream to upstream, or upstream to donstream 
as it is. As there is SSL connection in both end, no changes are required in the communication packets.


#### Downstream TLS + Upstream No-TLS:

In this configuration, the connection from client to Envoy will be SSL secured, but the connection from Envoy to upstream server
won't.
\
\
According to the MySQL connection protocol, there'll be a `Handshake Request` from the server at the beginning of the connection.
We'll need to send it to downstream as it is.\
Then there'll be `ssl request` packet from the downstream client. After receiving it from the filter-end, we'll need to first
call the `connection.startSecureTransport()` function to start the key exchange process with the client and make the downstream 
connection SSL secured. Then we need to return `FilterStatus::StopIteration` from the filter to stop this packet from sending 
to upstream. This is because we don't want the upstream connection to be SSL secured.\
After that, we'll receive the `Handshake Response` packet from the client. We'll send this packet to upstream, but we need
to make some changes in the packet before sending.\
As we may found from the MySQL connection protocol, every packet contains a 3 byte `packet length` information at the beginning
of each packet, followed by a 1 byte packet index. This is fixed for every MySQL packet.\
In this configuration we have TLS in the downstream connection but not in the upstream connection. And so, we didn't send the 
`ssl request` to the upstream. Thus, the next packet indices won't be the same for upstream and downstream. So, starting from the `Handshake Response`
packet, we need to decrease the index by 1 for every packet going to upstream from the downstream and increase the index by 1
for every packet going to downstream from the upstream. This will continute until the authentication is finished and the connection
is established successfully.\
Besides this, we need to change the `client_ssl` flag in `capabilities flags` contained by the `Handshake Response` packet.
This flag comes as `true` as we have TLS enabled in the downstream connection. But before sending it to upstream, we need to make
this flag false.


#### Downstream No-TLS + Upstream TLS:
In this configuration the downstream connection won't be SSl secured, but the upstream connection with the server will be 
SSL secured. 
\
\
According to the MySQL connection protocol, we need to send a `ssl request` packet to establish SSL secured connection with
the upstream server. But the in this configuration the downstream client won't send this packet as there's no SSl encryption
in the downstream connection.\
So we need to build the `ssl request` packet from the filter and send it to upstream before sending the `Handshake Response` packet that
comes from the downstream client.\
To build `ssl request` packet we need to know the information of supported `character_set` by the client. So, we wait until the
`Handshake Response` packet comes to the MySQL filter, and we retrive the `character_set` from that packet.\
Then we build the `ssl request` packet and send this packet to upstream instead of the `Handshake Response` packet came from the downstream.
But we reserve the `Handshake Response` packet to send in next.\
Before sending the `ssl_request` packet to upstream, we register a `bytesSentCallback` function in the upstream connection object.
The upstream connection object can't be accessed directly from the `MySQL_filter` or the `FilterManagerImpl`. It is only accessible
from the `tcp_proxy` filter.\
So we send a request from `mysql_proxy` to `tcp_proxy` via `FilterManagerImpl`, this hierarchy may seem a bit complex,
but we can find it simply iterating the functions one by one. The chain starts from `MySQL proxy` by calling the function
`read_allbacks_.writeDataToUpstreamTCPConnFromWriteCB(data, end_stream)`. This function chain will go to `tcp_proxy` from where it'll call
a function named `sendDatafromDataSentCallback` written in `TcpUpstream`, and this class has the upstream `connection` object.
In this function, we'll register a `dataSentCallback` function, and inside that function we'll first try to make the connection
SSL secured by calling `connection.StartSecureTransport()` and then send the reserved `Handshake Response` packet to upstream connection
by calling `connection.write()` function.\
Note that, before sending this `Handshake Response` packet, we need to increase its index by `1` as the upstream now has one increased 
packet indexing because of sending the extra `ssl request` packet. We need to maintain this for the next packets of connection and authentication
phase.\
Also, make the `client_ssl` flag `true` in the `Handshake Response` packet's capabilities flags.


## Conclusion

If any issue arises in the connection phase, please try to find the issue by checking all packets according to the MySQL connection protocol
mentioned in the top. Try to find which packet is causing the issue and what information is getting wrong.
