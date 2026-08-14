import os

import boto3

s3 = boto3.Session(
    aws_access_key_id=os.getenv('AWS_ACCESS_KEY_ID'),
    aws_secret_access_key=os.getenv('AWS_SECRET_ACCESS_KEY'),
    region_name=os.getenv('AWS_REGION')
).client('s3')

def put_html(body: str, key: str):
    print(f"[+] Uploading HTML {key}")
    try:
        s3.put_object(
            Body=body,
            Bucket=os.getenv(key='AWS_BUCKET', default=""),
            Key=key,
            ContentType='text/html',
        )
        print(f"[+] Uploaded HTML {key}")
    except Exception as e:
        print(f"[-] Upload HTML failed {e}")

def put_png(body: bytes, key: str):
    print(f"[ ] Uploading PNG {key}")
    try:
        s3.put_object(
            Body=body,
            Bucket=os.getenv(key='AWS_BUCKET', default=""),
            Key=key,
            ContentType='image/png',
        )
        print(f"[+] Uploaded PNG {key}")
    except Exception as e:
        print(f"[-] Upload PNG failed {e}")
