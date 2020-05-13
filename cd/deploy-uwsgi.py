import subprocess

def application(env, start_response):
    res = subprocess.run(['bash', 'deploy.sh'], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)

    if res.returncode == 0:
        code = '200 OK'
    else:
        code = '500 Error'
    start_response(code, [('Content-Type','text/plain')])
    result = res.stdout
    print ("RESPONSE:\n%s" % result)
    return [result]
