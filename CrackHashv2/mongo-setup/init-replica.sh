#!/bin/bash
sleep 5

echo "Initiating replica set 'rs0'..."
mongosh --host mongo1:27017 <<EOF
rs.initiate({
  _id: "rs0",
  members: [
    { _id: 0, host: "mongo1:27017" },
    { _id: 1, host: "mongo2:27017" },
    { _id: 2, host: "mongo3:27017" }
  ]
});
EOF

echo "Waiting for PRIMARY selection..."
sleep 5

echo "rs.status() output:"
mongosh --host mongo1:27017 --eval "rs.status()"

echo "Replica set init process completed!"
exit 0
