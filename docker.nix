# It is possible to build docker images ~without~ docker or dockerfiles.
# That is possible and completely fine, but using a Dockerfile is likely more
# familiar and less intimidating to people and because so many tools in the
# ecosystem often ingest and use Dockerfiles.
{ pkgs ? import <nixpkgs> {} }:

let
  # Import the Go package from default.nix
  urlshortener = import ./default.nix { inherit pkgs; };

  # Apply overrides to the Go package
  linuxArm64Shortener = urlshortener.overrideAttrs (oldAttrs: rec {
    nativeBuildInputs = oldAttrs.nativeBuildInputs or [] ++ [ pkgs.makeWrapper ];
    # Set GOOS, GOARCH, and CGO_ENABLED for cross-compilation
    preBuild = ''
      export GOOS=linux
      export GOARCH=arm64
      export CGO_ENABLED=0
    '';

    # Do not run tests. They fail on MacOS because we
    # use GOOS=linux. We should do it when building the image.
    doCheck = false;
  });

in pkgs.dockerTools.buildLayeredImage {
  name = "nixurlshortener";
  tag = "nix";
  contents = [ linuxArm64Shortener ];
  config = {
    #Cmd = [ "${urlshortener}/bin/urlshortener" ]; # Adjust the path if necessary
    Cmd = [ "${linuxArm64Shortener}/bin/linux_arm64/urlshortener" ]; # Adjust the path if necessary
    ExposedPorts = {
      "8080/tcp" = {};
    };
  };
}
