{ pkgs ? import <nixpkgs> {} }:

pkgs.buildGoModule rec {
  pname = "urlshortener";
  version = "0.1.0";

  # Specify the source directory.
  src = ./.;

  # Enable vendoring support.
  vendorHash = null;


  # Specify any system dependencies (for example, if you need Git).
  nativeBuildInputs = [ pkgs.pkg-config ];

  # Metadata for the package.
  meta = {
    description = "A simple URL shortener written in Go";
    homepage = "https://github.com/aevox/urlshortener";
    license = pkgs.lib.licenses.mit;
    maintainers = with pkgs.lib.maintainers; [ ];
  };
}
