# cryptocli

cryptocli is a command line tool to manage, control and access KeyControl Application Security Vault. 

The KeyControl Application Security Vault provides centralized mechanism to create and manage cryptographic keys and utilize those keys to encrypt, decrypt, tokenize, detokenize, sign, verify, wrap and unwrap data. 

cryptocli requires Entrust KeyControl version 10.2 or later. 

An Application Security Vault must be created by the KeyControl Security Administrator. 

Both Application Security Vault Users and Administrators can access secrets using the cryptocli with the login URL. 

Access to certain commands can be restricted using tokenization permissions from access policies.

## Releases

cryptocli's for Linux & Windows for each release can be found in Releases section (https://github.com/EntrustCorporation/cryptocli/releases)

## Build instructions

The code in this repo corresponds to the latest released version of cryptocli. In general, to use cryptocli, head over to Releases section to get pre-compiled binaries. If you do plan to build, follow instructions below.
1. Install go. cryptocli/Makefile expects to find go at /usr/local/go/bin/go
2. cd to cryptocli/
3. To build Linux & Windows cli binaries,
   
   ```$ gmake all```
5. To clean workspace,

   ```$ gmake clean```

For more information, see the Tokenization Vault chapter in the Key Management Systems documentation at https://trustedcare.entrust.com/.
