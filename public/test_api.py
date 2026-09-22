import urllib.request
import json

data = json.dumps({"username": "admin", "password": "password"}).encode("utf-8") # I don't know the password
req = urllib.request.Request("http://localhost:3002/api/login", data=data, headers={"Content-Type": "application/json"})
try:
    with urllib.request.urlopen(req) as response:
        res = json.loads(response.read().decode())
        token = res.get("token")
        
        req2 = urllib.request.Request("http://localhost:3002/api/dashboard", headers={"Authorization": "Bearer " + token})
        with urllib.request.urlopen(req2) as res2:
            print(res2.read().decode())
except urllib.error.HTTPError as e:
    print("HTTP Error:", e.code, e.read().decode())
except Exception as e:
    print("Error:", e)
