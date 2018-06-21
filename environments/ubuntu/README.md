## Ubuntu VM

Example:
Download the latest version of [Ubuntu Desktop](https://www.ubuntu.com/download/desktop) and create the Ubuntu VM with your preferred virtualization software.

## Install Docker

Follow the instructions [here](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-docker-ce) to install docker.

>Note:
>
>If you are seeing the following error after running `sudo apt-get install docker-ce` to install `docker-ce`.
>
>```
>E: Invalid operation docker-ce
>```
>
>Try running the following command to setup the **stable** repository that instead specifies an older Ubuntu distribution like `xenial` instead of using `lsb_release -cs` (using `bionic` doesn't seem to always works).
>
>```
>$ sudo add-apt-repository \
>   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
>   xenial \
>  stable"
>```
>
>Now try running `$ sudo apt-get install docker-ce` again.

To manage docker as a non-root user (so you don't have to keep using `sudo`) follow the instructions [here](https://docs.docker.com/install/linux/linux-postinstall/). **You must log out and log back in (or just restart your VM) so that your group membership is re-evaluated**
