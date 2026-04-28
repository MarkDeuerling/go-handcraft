# go-handcraft

## Upgrade Go version (Linux)
1. Download
``wget https://go.dev/dl/go1.26.0.linux-amd64.tar.gz``

2. Remove old
``rm -rf /usr/local/go``

3. Unzip Download and copy it to /usr/local
``tar -C /usr/local -xzf go1.26.0.linux-amd64.tar.gz``

### Set Path to go
edit ``.bashrc`` or ``~/.zshrc`` and add to PATH
``export PATH=$PATH:/usr/local/go/bin``

### Upgrade go mod in project
``go mod tidy -go=1.26``


