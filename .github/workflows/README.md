# AWS Profile Switching with Golang

![AWS Profile Switching](link_to_image)

## Table of Contents
- [Introduction](#introduction)
- [Installation](#installation)
- [Contributing](#contributing)
- [License](#license)

## Introduction

Managing multiple AWS profiles can be challenging, especially when working on different projects or environments. This Golang-based AWS Profile Switching tool aims to simplify the process of switching between AWS profiles, making it more efficient and error-free.

The code utilizes the official AWS SDK for Go to load AWS configuration and credentials, providing a seamless way to handle profile selection from the available options. This lightweight CLI tool is user-friendly and can be easily integrated into your existing development workflow.

## Installation

To get the latest release of the AWS Profile Switching tool, follow these steps:

1. Go to the [Releases](https://github.com/pooriaghaedi/aws-profile-switcher/releases) section of the GitHub repository.

2. Look for the latest release version (e.g., `v1.0.3`) and download the appropriate release for your operating system and architecture. For example, if you are on a macOS system with an AMD64 architecture, you might download `aws-profile-switcher_1.0.3_darwin_amd64.tar.gz`.

3. Move the downloaded release to `/usr/local/sbin`, using the following command (you may need administrative privileges):

   ```bash
   tar -xzvf aws-profile-switcher_1.0.3_darwin_amd64.tar.gz
   sudo mv /path/to/downloaded/aws-profile-switcher /usr/local/sbin/pfsw
   rm -f aws-profile-switcher_1.0.3_darwin_amd64.tar.gz
   pfsw

## Contributing

Contributions are welcome and encouraged! If you have any ideas, suggestions, or bug fixes, please feel free to open an issue or submit a pull request. Please make sure to adhere to the existing code style and add appropriate tests for any new functionalities.

## License

This project is licensed under the MIT License.