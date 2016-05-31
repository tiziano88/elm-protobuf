{ pkgs ? import <nixpkgs> {} }:

with pkgs;

stdenv.mkDerivation rec {
  name = "elm-protobuf";
  builder = "./builder.sh";
  inherit protobuf3_0;
  buildInputs = [ elm go protobuf3_0 ];
}
