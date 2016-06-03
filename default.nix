{ pkgs ? import <nixpkgs> {} }:

with pkgs;

stdenv.mkDerivation rec {
  name = "elm-protobuf";
  builder = "./builder.sh";
  buildInputs = [
    elmPackages.elm
    go
    goPackages.protobuf
    protobuf3_0
  ];
}
