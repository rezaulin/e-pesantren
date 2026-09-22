import mysql.connector
import urllib.request
import json
import bcrypt

# Connect to mysql to force update admin password
conn = mysql.connector.connect(user='root', password='', host='127.0.0.1', database='pesantren_multi')
cursor = conn.cursor()

# Hash a known password "testpass"
hashed = bcrypt.hashpw(b"testpass", bcrypt.gensalt())
cursor.execute("UPDATE users SET password_hash = %s WHERE username = 'admin'", (hashed,))
conn.commit()
print("Updated admin password to 'testpass'")

data = json.dumps({"username": "admin", "password": "testpass"}).encode("utf-8")
req = urllib.request.Request("http://localhost:3002/api/login", data=data, headers={"Content-Type": "application/json"})
try:
    with urllib.request.urlopen(req) as response:
        res = json.loads(response.read().decode())
        token = res.get("token")
        
        req2 = urllib.request.Request("http://localhost:3002/api/dashboard", headers={"Authorization": "Bearer " + token})
        with urllib.request.urlopen(req2) as res2:
            print("Dashboard success:", res2.read().decode())
except urllib.error.HTTPError as e:
    print("HTTP Error:", e.code, e.read().decode())
except Exception as e:
    print("Error:", e)
