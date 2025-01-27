{
  description = "A Nix-flake-based development environment for the regatta-watch monorepo";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
        forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
          pkgs = import nixpkgs { inherit system; };
        });
    in
    {
      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [

            # task runner
            just

            # frontend development tools
            nodejs_22
            pnpm

            # backend development tools
            go
            golangci-lint
            gotools

            # database
            goose
            postgresql
          ];
        };
      });
    };
}
