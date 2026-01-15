#!/bin/bash
cd /Users/playplus/goim/goim
nohup ./bin/chatapi cmd/chatapi/chatapi.toml > logs/chatapi.log 2>&1 &
echo $! > logs/chatapi.pid
sleep 3
curl -s http://localhost:3112/health
