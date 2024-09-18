{
  description = "secrets-manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/release-24.05";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    xc = {
      url = "github:joerdav/xc";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, gomod2nix, gitignore, xc }:
    let
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        inherit system;
        pkgs = import nixpkgs { inherit system; };
      });
    in
    {
      packages = forAllSystems ({ system, pkgs, ... }:
        let
          buildGoApplication = gomod2nix.legacyPackages.${system}.buildGoApplication;
        in
        rec {
          default = secrets-manager;

          secrets-manager = buildGoApplication {
            name = "secrets-manager";
            src = gitignore.lib.gitignoreSource ./.;
            go = pkgs.go;
            pwd = ./.;
            subPackages = [ "cmd/secrets-manager" ];
            CGO_ENABLED = 0;
            flags = [
              "-trimpath"
            ];
            ldflags = [
              "-s"
              "-w"
              "-extldflags -static"
            ];
            # Ensure the markdown file is included
            postBuild = ''
              cp ${./path/to/your/markdown/file.md} $out/
            '';
          };
        });

      devShell = forAllSystems ({ system, pkgs }:
        pkgs.mkShell {
          buildInputs = with pkgs; [
            golangci-lint
            cosign
            esbuild
            go_1_22
            gomod2nix.legacyPackages.${system}.gomod2nix
            gopls
            goreleaser
            gotestsum
            ko
            nodejs
            xc.packages.${system}.xc
          ];
        });

      overlays.default = final: prev: {
        secrets-manager = self.packages.${final.stdenv.system}.secrets-manager;
      };
    };
}