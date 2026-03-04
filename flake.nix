{
  description = "exarp-go development environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "aarch64-darwin" "x86_64-darwin" "aarch64-linux" "x86_64-linux" ];
      forEachSystem = nixpkgs.lib.genAttrs supportedSystems;
      pkgsFor = system: import nixpkgs { inherit system; };
    in
    {
      devShells = forEachSystem (system:
        let
          pkgs = pkgsFor system;
          # go.mod specifies go 1.25.5; use go from nixpkgs (24.11 has go_1_22).
          # Override with pkgs.go_1_23 or go_1_24 when available if needed.
          go = pkgs.go_1_22;
        in
        {
          default = pkgs.mkShell {
            name = "exarp-go-dev";
            buildInputs = with pkgs; [
              go
              gcc
              make
              git
              protobuf
              golangci-lint
            ];
            # Ensure Go modules and binaries are on PATH
            shellHook = ''
              export PATH="$PATH:$(dirname $(which go))/../share/go/bin"
              export CGO_ENABLED=1
              echo "exarp-go dev shell: go, make, git, protoc, golangci-lint"
            '';
          };
        });
    };
}
