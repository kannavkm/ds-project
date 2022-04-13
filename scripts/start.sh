#!/bin/bash

make server
./build/bin/server -p 8001 -cluster "127.0.0.1:8001,127.0.0.1:8001" 
