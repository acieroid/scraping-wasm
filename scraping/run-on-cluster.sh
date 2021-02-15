# This scripts runs the scraping on a cluster of multiple servers, assuming you
# can ssh into each of the server.
./make.sh || exit 1 # You can also make locally and scp files on the server if you don't have root access to install go etc..

# This is supposed to be launched from the main server, with the following IP.
# This is where the coordinator will live.
COORDINATOR_IP="10.0.0.10"
./docker/launch-coordinator "$COORDINATOR_IP:6345" || exit 1
sleep 60 # Make sure that it is launched

# These are all the nodes on which we will deploy workers
NODES="""
  10.0.0.10
  10.0.0.11
  10.0.0.12
  10.0.0.13
  10.0.0.14
  10.0.0.15
  10.0.0.16
  10.0.0.17
  10.0.0.18
"""
PWD=$(pwd) # We assume everything is mirorred on the other machines too for simplicity
for node in $NODES; do
    ssh -o StrictHostKeyChecking=no $node "cd $(PWD) && ./docker/launch-node $COORDINATOR_IP:6345 $NODE:6346"
done
