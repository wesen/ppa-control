#!/bin/sh

JQ_CMD=$(cat <<-EOM
.[]._source.layers
   | {
      dst_ip: .["ip.dst"][0],
      dst_port: .["udp.dstport"][0],
      src_ip: .["ip.src"][0],
      src_port: .["udp.srcport"][0],
      data: .data[0]
    }
   | [.src_ip, .src_port, .dst_ip, .dst_port, .data]
   | @tsv
EOM
)

tshark -r "$1"  \
    -Y "udp.port==5001" \
    -e data -e udp.srcport -e udp.dstport \
    -e ip.src -e ip.dst \
    -Tjson \
    | jq -r "$JQ_CMD"