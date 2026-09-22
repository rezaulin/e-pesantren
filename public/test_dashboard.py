import urllib.request

req = urllib.request.Request("http://localhost:3002/api/dashboard", headers={"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MSwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOiJhZG1pbiIsIm5hbWEiOiJBZG1pbiIsInRlbmFudF9pZCI6MSwiZXhwIjoxNzgxODY3NDUwfQ.MeAC-00rUdOn5lSFWkvSMuGcLs2TbnlY4ns-vASn5nQ"})
try:
    with urllib.request.urlopen(req) as res:
        print("Success:", res.read().decode())
except urllib.error.HTTPError as e:
    print("HTTP Error:", e.code, e.read().decode())
except Exception as e:
    print("Error:", e)
