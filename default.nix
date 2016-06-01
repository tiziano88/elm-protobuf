{ pkgs ? import <nixpkgs> {} }:

with pkgs;

stdenv.mkDerivation rec {
  name = "elm-protobuf";
  builder = "./builder.sh";
  inherit protobuf3_0;
  buildInputs = [
    elmPackages.elm
    go
    goPackages.protobuf
    protobuf3_0
  ];
}
