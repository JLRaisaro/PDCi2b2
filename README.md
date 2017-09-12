# PrivateDCi2b2 
PrivateDCi2b2 is a client-server platform for securely sharing aggregate-level i2b2 data in the cloud. PrivateDCi2b2 is based on additively homomorphic encryption and proxy re-encryption in order to provide end-to-end data-confidentiality protection. It's goal is to encourage clinical sites using i2b2 to share aggregate-level data on an untrusted public cloud such as AWS or Google Cloud.

PrivateDCi2b2 is developed by lca1 (Laboratory for Communications and Applications in EPFL) in collaboration with the ARCH team at HMS (Harvard Medical School).  

## Documentation

* The PrivateDCi2b2 platform does an intensive use of [Overlay-network (ONet) library](https://github.com/dedis/onet)
* For more information regarding the underlying architecture please refer to the stable version of ONet `gopkg.in/dedis/onet.v1`
* To check the code organisation, have a look at [Layout](https://github.com/lca1/unlynx/wiki/Layout)
* For more information on how to run PrivateDCi2b2, simulations and apps, go to [Running UnLynx](https://github.com/lca1/unlynx/wiki/Running-UnLynx)

## Getting Started

To use the code of this repository you need to:

- Install [Golang](https://golang.org/doc/install)
- [Recommended] Install [IntelliJ IDEA](https://www.jetbrains.com/idea/) and the GO plugin
- Set [`$GOPATH`](https://golang.org/doc/code.html#GOPATH) to point to your workspace directory
- Add `$GOPATH/bin` to `$PATH`
- Git clone this repository to $GOPATH/src `git clone https://github.com/lca1/unlynx.git` or...
- go get repository: `go get github.com/lca1/unlynx`

## Version

The version in the `master`-branch is stable and has no incompatible changes.

## License

UnLynx is licensed under a End User Software License Agreement ('EULA') for non-commercial use. If you want to have more information, please contact us.

## Contact
You can contact any of the developers for more information or any other member of [lca1](http://lca.epfl.ch/people/lca1/):

* [David Froelicher](https://github.com/froelich) (PHD student) - david.froelicher@epfl.ch
* [Patricia Egger](https://github.com/pegger) (Security Consultant at Deloitte) - patricia.egger@epfl.ch
* [Joao Andre Sa](https://github.com/JoaoAndreSa) (Software Engineer) - joao.gomesdesaesousa@epfl.ch
* [Christian Mouchet](https://github.com/ChristianMct) (MSC student) - christian.mouchet@epfl.ch
