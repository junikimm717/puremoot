{
  description = "Clean Permoot";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-23.05";
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
  };

  outputs = { self, nixpkgs, flake-utils }:
    with flake-utils.lib;
    eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        version = "0.0.0";
        name = "pureMOOt";
      in {

        devShell = pkgs.mkShell { buildInputs = with pkgs; [ go gopls scc redis ]; };

        defaultPackage = pkgs.stdenv.mkDerivation { inherit name version; };

        formatter = nixpkgs.legacyPackages."${system}".nixfmt;

      });
}
