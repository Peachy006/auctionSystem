
go build -o auction-node.exe ./node
go build -o auction-client.exe ./client

PEERS=5000,5001,5002,5003,5004,5005,5006,5007

trap 'kill 0' EXIT

for port in {5000..5007}; do
  ./auction-node.exe -port $port -peers $PEERS &
done

sleep 1

for i in {1..4}; do
  echo "client$i" | ./auction-client.exe &
done

wait