using Go = import "/go.capnp";
@0x9c1e5a2b3d4f6a7b;
$Go.package("bldrec");
$Go.import("bldrec");

struct Record {
  uDateTime   @0 :Int64;      # std::time_t is typically a 64‑bit integer
  sDescription @1 :Text;      # std::string → Text
  fValue      @2 :Float32;    # float → Float32
  sDontCare   @3 :Text;       # std::string → Text
}
