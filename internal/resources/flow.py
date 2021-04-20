import requests

print('::{"outputs": {"extract":"' + str(requests.get('http://google.com').status_code) + '"}}::')
