{ pkgs ? import <nixpkgs> {} }:

with pkgs;
with goPackages;

goPackages.buildGoPackage rec {
  rev = "1111461c35931a806efe06a9a43ad52a24c608ff";
  name = "goproto";
  goPackagePath = "github.com/golang/protobuf";
  src = fetchFromGitHub {
    inherit rev;
    owner = "golang";
    repo = "protobuf";
    sha256 = "12py0r0vfn5gb2cgzwd3bqi8lqg1jgp805nw7md8fb44sl5qsyvg";
  };
}
