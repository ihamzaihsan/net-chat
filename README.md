# net-cat

## Description
This project consists on recreating the NetCat in a Server-Client Architecture that can run in a server mode on a specified port listening for incoming connections, and it can be used in client mode, trying to connect to a specified port and transmitting information to the server.

NetCat, nc system command, is a command-line utility that reads and writes data across network connections using TCP or UDP. It is used for anything involving TCP, UDP, or UNIX-domain sockets, it is able to open TCP connections, send UDP packages, listen on arbitrary TCP and UDP ports and many more.

## Authors
- Yousif Maidan(ymaidan)
- hamza cheema(hcheema)

## Important notes
1. Control connections quantity (Maximum 10 connections).    
2. TCP connection between server and multiple clients (relation of 1 to many).  
3. A name requirement to the client.  
4. Clients must be able to send messages to the chat.  
5. Do not broadcast EMPTY messages from a client.  
6. Messages sent, must be identified by the time that was sent and the user name of who sent the message, example : [2020-01-20 15:48:41][client.name]:[client.message]  
7. If a Client joins the chat, all the previous messages sent to the chat must be uploaded to the new Client.  
8. If a Client connects to the server, the rest of the Clients must be informed by the server that the Client joined the group (<name> has joined our chat... ).  
9. If a Client exits the chat, the rest of the Clients must be informed by the server that the Client left (<name> has left our chat...).  
10. All Clients must receive the messages sent by other Clients.  
11. If a Client leaves the chat, the rest of the Clients must not disconnect.  
12. If there is no port specified, then set as default the port 8989. Otherwise, program must respond with usage message: [USAGE]: ./TCPChat $port  

## Usage
### Running the program
1. go run .                  --> default port 8989  
2. go run . <port>           --> specified port  
3. go run . <port> localhost --> [USAGE]: ./TCPChat $port  

### After running the program the client must specify the ip and port
<pre>
1. nc $IP $port  

2. The following must be printed  
Welcome to TCP-Chat!  
         _nnnn_  
        dGGGGMMb  
       @p~qp~~qMb  
       M|@||@) M|  
       @,----.JM|  
      JS^\__/  qKL  
     dZP        qKRb  
    dZP          qKKb  
   fZP            SMMb  
   HZM            MMMM  
   FqM            MMMM  
 __| ".        |\dS"qML  
 |    `.       | `' \Zq  
_)      \.___.,|     .'  
\____   )MMMMMP|   .'  
     `-'       `--'  
[ENTER YOUR NAME]: $name 

</pre>
# Sending a message format
[2020-01-20 15:48:41][client.name]:[client.message] 

