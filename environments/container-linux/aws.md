# AWS instance


```bash
./ignition.py < ignition.yaml > ignition.json
```

Use the CoreOS [validate tool](https://coreos.com/validate)

Mind to replace:
* image-id
* key-name
* region
```bash
aws ec2 run-instances \
    --image-id "ami-5d6e5e38" \
    --instance-type "t2.large" \
    --count 1 \
    --key-name "julien.balestra" \
    --region "us-east-2" \
    --user-data file://ignition.json \
    --instance-initiated-shutdown-behavior "terminate" \
    --block-device-mapping '{"DeviceName":"/dev/xvda","Ebs":{"DeleteOnTermination":true,"VolumeSize": 15,"VolumeType":"gp2"}}'
```
