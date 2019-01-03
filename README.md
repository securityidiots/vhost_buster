# vhost_buster
A simple tool by securitydiots with the power of "Go" to find the hidden Vhosts defined at the server. Many times there are hidden virtual hosts defined at server side, without a DNS entry. Such hosts could not be directly found because our machine will query the DNS server to resolve the host to an IP, since the DNS record is not there, it will fail. 

Using this tool we can resolve our host to a particular server IP. 

### Why
1. There is no other good option to look for virtual hosts.
2. I Decided to check out Golang and do a practice project.
3. The community have taugh me a lot, I thought may be I can do some good for others.

### How to use
Use ```vhost_buster -h``` to see help.

### Example
![example image](https://github.com/securityidiots/vhost_buster/blob/master/Screenshot.png?raw=true)

It will show the first result infinityworriortest as the base result of the server response. If the response change next time, it will again show you the virtual host in output.


### ToDo List
1. Create a Todo list :P.
