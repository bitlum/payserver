from flask import Flask
from flask import request
from subprocess import Popen, PIPE

app = Flask(__name__)

rpcAddr = "bitcoin-lightning.simnet.primary:10009"

macaroonFile = "/root/.lnd/data/chain/bitcoin/regtest/admin.macaroon"
tlsCertFile = "/root/.lnd/tls.cert"

@app.route('/pay_invoice')
def pay_invoice():
    invoice = request.args.get('invoice')

    bashCommand = "lncli --network=regtest --rpcserver={} --macaroonpath={} " \
                  "--tlscertpath={} payinvoice --force --pay_req={}".format( \
                  rpcAddr, macaroonFile, tlsCertFile, invoice)

    print(bashCommand)
    p = Popen(bashCommand, shell=True, stdout=PIPE)
    out = p.communicate()[0]
    return str(out)

@app.route('/generate_invoice')
def generate_invoice():
    amount = request.args.get('amount')

    bashCommand = "lncli --network=regtest --rpcserver={} --macaroonpath={} " \
                  "--tlscertpath={} addinvoice --amt={}".format( \
                  rpcAddr, macaroonFile, tlsCertFile, amount)

    print(bashCommand)
    p = Popen(bashCommand, shell=True, stdout=PIPE)
    out = p.communicate()[0]
    return str(out)

if __name__ == '__main__':
    app.run(host="0.0.0.0", port="80")


