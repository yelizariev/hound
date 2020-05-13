#!/usr/bin/env python3
import subprocess

def application(env, start_response):
    print("REQUEST: %s", env.get('REQUEST_URI'))
    res = subprocess.run(['bash', 'deploy.sh', env.get('REQUEST_URI')], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)

    if res.returncode == 0:
        code = '200 OK'
    else:
        code = '500 Error'
    start_response(code, [('Content-Type','text/plain')])
    result = res.stdout
    print ("RESPONSE:\n%s" % result)
    return [result]
