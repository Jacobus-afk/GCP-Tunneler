{
  description = "A Nix-flake-based Python development environment";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";

  outputs =
    { nixpkgs, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { system = "${system}"; };
    in
    {
      devShells.${system} = {
        default = pkgs.mkShell {
          packages =
            with pkgs;
            [
              go
              gopls
              gotools
              gofumpt
              gomodifytags
              impl
              delve
              golines
            ];
        };
      };
    };
}
