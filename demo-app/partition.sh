#!/bin/bash
sudo iptables-save > ~/iptables.save
sudo iptables -A INPUT -p tcp --destination-port 10000 -j DROP
sudo iptables -A OUTPUT -p tcp --destination-port 9999 -j DROP
sudo iptables -A OUTPUT -p tcp --destination-port 10000 -j DROP
