# PINGER

NOTE: COMMIT HISTORY is only one, as it was in sync with the [old repository](https://github.com/NetworkInCode/networkincode-classroom-custom-ping-utility-custom_ping_utility). To see the full commit history, please refer tho this GitHub repository: https://github.com/NetworkInCode/custom-ping-utility-Vishy70

`pinger` is a custom **ping(8)** clone, to send ICMP ECHO_REQUEST to network hosts on the internet, built in Golang!

## Instructions

Currently, ther is no readily available versioned binary for `pinger`, so you need to compile from source. It currently only supoorts Linux and MacOS.

### Prerequisites

You need to have the Go toolchain set up on your machine.
The [download and install](https://go.dev/doc/install) page on [Go's documentation](https://go.dev/doc) is a good place to start. 
A summary of it follows.

#### Linux

- On the [website](https://go.dev/doc/install), click the download button and choose the Linux .tar.gz archive file.

- Open a shell/terminal and navigate to locations where the archive downloaded (by default, it would be in Dowloads so run `cd ~/Downloads`).

- Run the following (you may need to run as root, or use sudo):

    `rm -rf /usr/local/go && tar -C /usr/local -xzf go<version-number-here>.linux-amd64.tar.gz`

- Add the `user/local/go/bin` directory to your `$PATH` shell environment variable. 
To do this globally, you need to modfiy your `$HOME/.profile`. You can do this by opening the file (`nano $HOME/.profile`), and adding the following line on a new line at the end of the file: `export PATH=$PATH:/usr/local/go/bin`. Save and exit the file.

- Verify that Go is now installed correctly by running `go version`. It should show something like:

    `go version go1.24.1 linux/amd64`

- NOTE: You may need to open a new terminal window, for the changes in the `$HOME/.profile` to take place.

- You can now build the `pinger` package!

#### MacOS

- On the [website](https://go.dev/doc/install), click the download button and choose the Apple macOS (either ARM64 x86_64 depending on CPU architecture) .pkg file. To check this, you can open a terminal and type the `uname -p` command to determine it.

- Double click and open the package you installed, and follow the wizard instructions. The installer should install the package in `/usr/local/go`, and put `/usr/local/go/bin` in your PATH environment variable.

- You can now build the `pinger` package!

### Dependencies

The exact dependencies for the project can be found in the go.mod file in the `pinger` directory. The dependencies will get installed when running `go build`, as below.

### Build and Install Pinger

To get build and install the binary, do the following :-

- Clone the repository in a directory, and enter the project directory:

    `git clone https://github.com/NetworkInCode/custom-ping-utility-Vishy70 && cd custom-ping-utility-Vishy70`

- Change into the `pinger` directory: 

    `cd pinger`

- Build the application (you can change the destination by modifying the path given to *flag -o*): 

    `go build -o ../bin/pinger`

- You can now run the application (NOTE: You will likely need root access, or need to run the application suing sudo):

    `../bin/pinger --help`

- You can also install it as a binary that can be run as `pinger -h` by running `go install` inside of pinger directory, although not suggested.
    Note: You may likely get permission error, it is likely that your $GOBIN is pointing to /bin, which is not writable by non-root. You will need to setup your $GOBIN and $PATH root environment variables to `/bin` and `$PATH:/usr/local/go/bin`. Then, as root, you can directly use `pinger` command itself. Thus, it is advised to use `go build` instead, as you can easily run that with sudo. 

## Usage

`pinger` sends ICMP Echo Requests. There is only one required argument: a destination address. This can be an IPv4 address, IPv6 address, or even a internet host-name (it will automatically get resolved via DNS), as long as the address is not malformed.
    `./pinger nitk.ac.in`

In addition, there are some `flags` that can modify `pinger`'s functionality:-
- Use [-4|-6] to specifically use an IPv4/IPv6 address. These are mutually exclusive flags.
- Use [-I] <iface-name> to specify the network device you want to send and receive ICMP Echo Requests and Replies from.
- Use [-c] <number-of-times> to specify the number of Echo Requests you want to send
- Use [-t ] <ttl> to set the packet Time To Live 

An example: `./pinger -I wlp45s0 -c 4 -6 nitk.ac.in`

## Scripts

There is a script to run an example usage of pinger in [`scripts/test.sh`](./scripts/test.sh). This script was made with Linux in mind, some minor modifcations may be required for MacOS. 

IMPORTANT: the script expects an absolute path to the packages `pinger directory`. Please keep this in mind!

## Features
- Mostly compatible with both **Linux and macOS systems**.
- **Error handling** to deal with network timeouts, unreachable hosts, etc.
- IPv4/IPv6 support included
- Custom flags for **network interface**, **number of echo requests**, **ttl**.

## Issues
- Unable to handle IPv6's neighbor advertisement & solicitation, and router advertisement & solicitation. Does not count these errors...
- Exact ttl value in packet won't be set **(similar to ping(8))**, but will still account for TTL Exceeded, and Hop Limit Reached successfully.
- Since requires root privileges, the user currently needs to do some extra configuration of GO Environment variables (for root user) to run the binary simply as `pinger`.