{ pkgs ? import <nixpkgs> {} }:

let
  goEnv = pkgs.go;
  vscodeExtensions = [
    pkgs.vscode-extensions.golang.go
  ];
in
pkgs.mkShell {
  buildInputs = [
    goEnv
    pkgs.vscode
  ] ++ vscodeExtensions;

  shellHook = ''
    export GOPATH=$HOME/go
    export PATH=$GOPATH/bin:$PATH
    echo "Go environment is set up. GOPATH is $GOPATH"
  '';
}
