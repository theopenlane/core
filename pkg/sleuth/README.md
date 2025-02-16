# Sleuth

Package sleuth is a set of tools for performing DNS enumeration, http probing, and other reconnaissance tasks. It is designed to be used as a library and contains wrappers around several [Project Discovery](https://github.com/projectdiscovery) for ease of use as a library.

## Overview

Slueth consists of several subparts:

- **dnsx**: DNS record querying and analysis along with subdomain enumeration
- **httpx**: HTTP probing and analysis
- **ports**: Port scanning using naabu and nmap
- **emailx**: Email address analysis, confirmation of conmpromise or leakage
- **spider**: Web crawling and analysis
- **certx**: Certificate analysis

## Prerequisites

- Go 1.23 or later
- [Project Discovery](https://github.com/projectdiscovery)

## Installation

## DNS

### What is DNS bruteforcing?

DNS bruteforcing is a technique used to discover subdomains of a target domain by appending a list of common subdomain names to the target domain and attempting to resolve them. This method is often used in penetration testing and security assessments to identify potential attack vectors or misconfigured subdomains.

This is what happens during DNS bruteforcing:

admin            ---->       admin.example.com

internal.dev  ---->       internal.dev.example.com

secret            ---->       secret.example.com

backup01      ---->       backup01.example.com

Now that we have a list of probable domain names that could exists, we can perform DNS resolution on this domain list. This would yield us live subdomains. After this process, if any of these subdomains is found valid, it's a win-win situation for us.

### Why do we perform subdomain bruteforcing?

At times passive DNS data doesn't give all the hosts/subdomains associated with our target. Also, there would some newer subdomains that still wouldn't have been crawled by the internet crawlers. In such a case subdomain bruteforcing proves beneficial.

Earlier DNS zone transfer vulnerabilities were the key to get the whole DNS zone data of a particular organization. But lately, the DNS servers have been secured and zone transfers are found very rarely.

### Wordlists

The whole effort of DNS bruteforcing is a waste if you don't use a good subdomain bruteforcing wordlist. Selection of the wordlist is the most important aspect of bruteforcing. Let's have a look at some great wordlists:-

1) Assetnote best-dns-wordlist.txt (9 Million) ⭐
Assetnote wordlists are the best. No doubt this is the best subdomain bruteforcing wordlist. But highly recommended that you run this in your VPS. Running on a home system will take hours also the results wouldn't be accurate. This wordlist will definitely give you those hidden subdomains.

2) n0kovo n0kovo_subdomains_huge.txt (3 Million)
N0kovo created this wordlist by scanning the whole IPv4 and collecting all the subdomain names from the TLS certificates. You can check out this blog to see how good this bruteforcing wordlist performs as compared other big wordlists. So, if you are target contains a lot of wildcards this would be best wordlist for bruteforcing(considering the computation bottleneck for wildcard filtering).

3) Smaller wordlist (102k )
Created by six2dez is suitable to be run if you are using your personal computer which is consuming your home wifi router internet.