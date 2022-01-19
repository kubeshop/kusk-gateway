######################################################################
# Most of these commands are borrowed from FreeRADIUS certs management (/etc/freeradius/certs)
#
######################################################################

DH_KEY_SIZE	= 2048
OPENSSL		= openssl
EXTERNAL_CA	= $(wildcard external_ca.*)

ifneq "$(EXTERNAL_CA)" ""
PARTIAL		= -partial_chain
endif

#
#  Set the passwords
#
include passwords.mk

######################################################################
#
#  Make the necessary files, but not client certificates.
#
######################################################################
.PHONY: all
all: index.txt serial dh ca server

.PHONY: client
client: client.pem

.PHONY: ca
ca: ca.der ca.crl

.PHONY: server
server: server.pem server.vrfy

.PHONY: verify
verify: server.vrfy client.vrfy

passwords.mk: server.cnf ca.cnf client.cnf
	@echo "PASSWORD_SERVER = '$(shell grep output_password server.cnf | sed 's/.*=//;s/^ *//')'"		> $@
	@echo "PASSWORD_CA	= '$(shell grep output_password ca.cnf | sed 's/.*=//;s/^ *//')'"		>> $@
	@echo "PASSWORD_CLIENT	= '$(shell grep output_password client.cnf | sed 's/.*=//;s/^ *//')'"		>> $@
	@echo "USER_NAME	= '$(shell grep emailAddress client.cnf | grep '@' | sed 's/.*=//;s/^ *//')'"	>> $@
	@echo "CA_DEFAULT_DAYS  = '$(shell grep default_days ca.cnf | sed 's/.*=//;s/^ *//')'"			>> $@

######################################################################
#
#  Diffie-Hellman parameters
#
######################################################################
dh:
	$(OPENSSL) dhparam -out dh -2 $(DH_KEY_SIZE)

######################################################################
#
#  Create a new self-signed CA certificate
#
######################################################################
ca.key ca.pem: ca.cnf
	@[ -f index.txt ] || $(MAKE) index.txt
	@[ -f serial ] || $(MAKE) serial
	$(OPENSSL) req -new -x509 -keyout ca.key -out ca.pem \
		-days $(CA_DEFAULT_DAYS) -config ./ca.cnf \
		-passin pass:$(PASSWORD_CA) -passout pass:$(PASSWORD_CA)
	chmod g+r ca.key

ca.der: ca.pem
	$(OPENSSL) x509 -inform PEM -outform DER -in ca.pem -out ca.der

ca.crl: ca.pem
	$(OPENSSL) ca -gencrl -keyfile ca.key -cert ca.pem -config ./ca.cnf -out ca-crl.pem -key $(PASSWORD_CA)
	$(OPENSSL) crl -in ca-crl.pem -outform der -out ca.crl
	rm ca-crl.pem

######################################################################
#
#  Create a new server certificate, signed by the above CA.
#
######################################################################
# NOTE '-nodes' means: do not encrypt the server.key
# Remove if need it encrypted
server.csr server.key: server.cnf
	$(OPENSSL) req -new -nodes -out server.csr -keyout server.key -config ./server.cnf
	chmod g+r server.key

server.crt: server.csr ca.key ca.pem
	$(OPENSSL) ca -batch -extensions x509_extensions -extfile ./server.cnf -keyfile ca.key -cert ca.pem -in server.csr -key $(PASSWORD_CA) -out server.crt -config ./server.cnf

server.p12: server.crt
	$(OPENSSL) pkcs12 -export -in server.crt -inkey server.key -out server.p12  -passin pass:$(PASSWORD_SERVER) -passout pass:$(PASSWORD_SERVER)
	chmod g+r server.p12

server.pem: server.p12
	$(OPENSSL) pkcs12 -in server.p12 -out server.pem -passin pass:$(PASSWORD_SERVER) -passout pass:$(PASSWORD_SERVER)
	chmod g+r server.pem

.PHONY: server.vrfy
server.vrfy: ca.pem
	@$(OPENSSL) verify $(PARTIAL) -CAfile ca.pem server.pem


######################################################################
#
#  Create a new client certificate, signed by the the above CA certificate.
#
######################################################################
client.csr client.key: client.cnf
	$(OPENSSL) req -new  -out client.csr -keyout client.key -config ./client.cnf
	chmod g+r client.key

client.crt: client.csr ca.pem ca.key
	$(OPENSSL) ca -batch -keyfile ca.key -cert ca.pem -in client.csr  -key $(PASSWORD_CA) -out client.crt -config ./client.cnf

client.p12: client.crt
	$(OPENSSL) pkcs12 -export -in client.crt -inkey client.key -out client.p12  -passin pass:$(PASSWORD_CLIENT) -passout pass:$(PASSWORD_CLIENT)
	chmod g+r client.p12

client.pem: client.p12
	$(OPENSSL) pkcs12 -in client.p12 -out client.pem -passin pass:$(PASSWORD_CLIENT) -passout pass:$(PASSWORD_CLIENT)
	chmod g+r client.pem
	cp client.pem $(USER_NAME).pem

.PHONY: client.vrfy
client.vrfy: ca.pem client.pem
	$(OPENSSL) verify -CAfile ca.pem client.pem


######################################################################
#
#  Miscellaneous rules.
#
######################################################################
index.txt:
	@touch index.txt

serial:
	@echo '01' > serial

print:
	$(OPENSSL) x509 -text -in server.crt

printca:
	$(OPENSSL) x509 -text -in ca.pem

clean:
	@rm -f *~ *old server.csr server.key server.crt server.p12 server.pem client.csr client.key client.crt client.p12 client.pem

#
#	Make a target that people won't run too often.
#
destroycerts:
	rm -f *~ dh *.csr *.crt *.p12 *.der *.pem *.key index.txt* \
			serial*  *\.0 *\.1 ca-crl.pem ca.crl
