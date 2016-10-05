rm -rf out
go get -v github.com/square/certstrap
ca_name=diegoCA
server_name=cell.service.cf.internal
server_domain='*.cell.service.cf.internal'
client_name='rep.client'


#  Initialize certificate authority
certstrap init --common-name $ca_name --passphrase ''

# Request and sign server certificate
certstrap request-cert --common-name $server_name --domain $server_domain --passphrase ''
certstrap sign $server_name --CA $ca_name

# Request and sign client certificate
certstrap request-cert --passphrase '' --common-name $client_name
certstrap sign $client_name --CA $ca_name

mv -f out/diegoCA.crt src/code.cloudfoundry.org/rep/cmd/rep/fixtures/green-certs/server-ca.crt
mv -f out/cell.service.cf.internal.crt src/code.cloudfoundry.org/rep/cmd/rep/fixtures/green-certs/server.crt
mv -f out/cell.service.cf.internal.key src/code.cloudfoundry.org/rep/cmd/rep/fixtures/green-certs/server.key
