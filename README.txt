To Run the Application:

Open 5 Terminal Windows
Enter the following commands for each of the windows

Window 1: go run appServer.go
Window 2: go run appServer2.go
Window 3: go run appServer3.go

Window 4: go run appClient.go
Window 5: go run appClient2.go

Enter a username in appClient2.go
Start typing messages in appClient2.go

The messages should go through the servers and return in appClient.go



If you want to connect the chat application to the EC2 servers, open 3 terminals.
In chatClient/
go run appTokyoClient.go
go run appEC2OrigClient.go
go run appSeoulClient.go

Type something in the windows for appSeoulClient and appTokyoClient, the messages will be received by appEC2OrigClient (Oregon).
