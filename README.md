## How to run the program

The program can be run in the traditional sense by running the nodes and running the clients.

For the node, it has 2 flags, the 2 flags are the port, and the other ports of the peers, for the peers it should just be seperated by commas.

The client has no flags and has the hardcoded adresses of the nodes. The hardcoded values are:
5000, 5001, 5002, 5003, 5004, 5005, 5006, 5007

## Running the program (Easy way)

We have created 2 bash scripts, one for running the program with no faults, it creates 8 nodes and 4 clients, and simulates the auction process until lamport 200.

The other bash script is called run-crash.sh and is designed to crash the used node of the cluster. It starts the program with 8 nodes and 4 clients, after 10 seconds it kills the process of the first node. Thus forcing the clients to reroute to the second node.