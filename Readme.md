# assume_role_with_mfa

A small MFA GUI tool designed to be used as `credential_proces` for AWS CLI configuration profiles.

This is a workaround for the official AWS MFA setup, that works for the CLI but is apparently broken for the Go SDK, as reported at [aws-sdk-go-v2/#2356](https://github.com/aws/aws-sdk-go-v2/issues/2356).

```text
[profile broken-sdk]
source_profile = users
role_arn=arn:aws:iam::bbbbbbbbbbbb:role/MyAssumedRole
mfa_serial = arn:aws:iam::aaaaaaaaaaaaa:mfa/my-mfa
```

Instead, we integrate as a credential_process, which is better supported by the Go SDK.

## Features

- It actually works for the Go SDK :-D
- Ask for MFA using a graphical user interface.
- Caches credentials locally for subsequent access.

## Prerequisites

- Go (version 1.15 or later)
- Valid AWS credentials configured as a profile in the AWS CLI credentials file.
- An AWS account configured with an MFA device

## Installation

Assuming you have Go set up and `$PATH` set to `$GOPATH/bin`, you will need to run the following command:

```shell
go install github.com/LeanerCloud/assume_role_with_mfa@latest
```

## Usage

Create this configuration in the `~/.aws/config` file. Make sure the path is correct. You can also test the command in a shell to see how it works. 

```text
[profile my_mfa]
credential_process = /Users/USERNAME/go/bin/assume_role_with_mfa -mfa-arn arn:aws:iam::XXXXXXXXXXXX:mfa/mfa -profile another-profile-having-static-credentials -role-arn arn:aws:iam::YYYYYYYYYYYY:role/myRole
region = MY_REGION
```

Use this new profile as usual and you will be asked for the MFA code using this small GUI.

<img width="181" alt="Screenshot 2023-12-05 at 11 15 13" src="https://github.com/LeanerCloud/assume_role_with_mfa/assets/95209/80aa3c5d-a485-40a9-919a-ba0068bedf5b">

The GUI will then assume the role using the MFA code and provide some temporary credentials to that profile.

## Credential caching

To avoid asking for MFA repeatedly, we cache the obtained credentials in a file located in the Fyne application's storage directory, and the cached credentials will be reused until they expire.

The cache filename is a SHA256 hash of the role ARN to uniquely identify the credentials.

## Related Projects

Check out our other open-source [projects](https://github.com/LeanerCloud)

- [awesome-finops](https://github.com/LeanerCloud/awesome-finops) - a more up-to-date and complete fork of [jmfontaine/awesome-finops](https://github.com/jmfontaine/awesome-finops).
- [Savings Estimator](https://github.com/LeanerCloud/savings-estimator) - estimate Spot savings for ASGs.
- [AutoSpotting](https://github.com/LeanerCloud/AutoSpotting) - convert On-Demand ASGs to Spot without config changes, automated divesification, and failover to On-Demand.
- [EBS Optimizer](https://github.com/LeanerCloud/EBSOptimizer) - automatically convert EBS volumes to GP3.
- [ec2-instances-info](https://github.com/LeanerCloud/ec2-instances-info) - Golang library for specs and pricing information about AWS EC2 instances based on the data from [ec2instances.info](https://ec2instances.info).
- [ipv4-cost-viewer](https://github.com/LeanerCloud/aws-ipv4-cost-viewer) - shows the future public IPv4 costs for a variety of AWS resources across all AWS regions from an account in a user-friendly terminal UI.

For more advanced features of some of these tools, as well as comprehensive cost optimization services focused on AWS, visit our commercial offerings at [LeanerCloud.com](https://www.LeanerCloud.com).

## Support

Reach out to us on [Slack](https://join.slack.com/t/leanercloud/shared_invite/zt-xodcoi9j-1IcxNozXx1OW0gh_N08sjg) if you need help or have any questions about this or any of our projects.

## Contributing

Contributions to this project are welcome! You can contribute in the following ways:

Report Issues: If you find any bugs or have feature suggestions, please create an issue.
Submit Pull Requests: Feel free to fork the repository and submit pull requests with bug fixes or new features.

## License

This project is licensed under the MIT License.
Copyright (c) 2023 Cristian Magherusan-Stanciu, [LeanerCloud.com](https://LeanerCloud.com).
