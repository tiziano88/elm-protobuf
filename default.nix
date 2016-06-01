{ pkgs ? import <nixpkgs> {} }:

with pkgs;

let
  goproto = callPackage ./goproto.nix {};
in
  stdenv.mkDerivation rec {
    name = "elm-protobuf";
    builder = "./builder.sh";
    inherit protobuf3_0;
    buildInputs = [
      elmPackages.elm
      go
      goproto
      protobuf3_0
    ];
  }
