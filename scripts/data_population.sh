#!/bin/bash

SERVICE_NAME="explore.ExploreService"
PROTO_FILE="../proto/explore-service.proto"
RECIPIENT_USER_ID="endy"
LIKED_RECIPIENT=true

# Loop to create 500 users
for i in $(seq 1 500); do
  ACTOR_USER_ID=$(uuidgen)

  PAYLOAD=$(cat <<EOF
{
  "actor_user_id": "$ACTOR_USER_ID",
  "recipient_user_id": "$RECIPIENT_USER_ID",
  "liked_recipient": $LIKED_RECIPIENT
}
EOF
)

  # Make the gRPC call using grpcurl
  grpcurl -plaintext \
    -proto "$PROTO_FILE" \
    -d "$PAYLOAD" \
    localhost:50051 $SERVICE_NAME/PutDecision

  echo "[$i] Actor User ID: $ACTOR_USER_ID"
  echo "-----------------------------"
done