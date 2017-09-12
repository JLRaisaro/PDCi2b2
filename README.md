# PDCi2b2 
PDCi2b2 (Private Data Characterization for i2b2) is a client-server platform for the privacy-preserving sharing of aggregate-level i2b2 data. PDCi2b2 is based on additively homomorphic encryption and proxy re-encryption. Its goal is to make clinical sites more comfortable with sharing data in an untrusted public cloud, such as AWS or Google Cloud, by providing strong end-to-end confidentiality protection. With PDCi2b2, data in the cloud are protected against a honest-but-curious adversary not aonly at rest and in transit but also during computations.

PDCi2b2 is developed by LCA1 (Laboratory for Communications and Applications in EPFL) in collaboration with the ARCH team at HMS (Harvard Medical School).  

## Documentation

* PDCi2b2 does an intensive use of [Overlay-network (ONet) library](https://github.com/dedis/onet)
* For more information regarding the underlying architecture please refer to the stable version of ONet `gopkg.in/dedis/onet.v1`
* To check the code organisation, have a look at [Layout](https://github.com/lca1/unlynx/wiki/Layout)
* For more information on how to run PDCi2b2, simulations and apps, go to [Running UnLynx](https://github.com/JLRgithub/PrivateDCi2b2/wiki/Running-PDCi2b2)

## Getting Started

To use the code of this repository you need to:

- Install [Golang](https://golang.org/doc/install)
- [Recommended] Install [IntelliJ IDEA](https://www.jetbrains.com/idea/) and the GO plugin
- Set [`$GOPATH`](https://golang.org/doc/code.html#GOPATH) to point to your workspace directory
- Add `$GOPATH/bin` to `$PATH`
- Git clone this repository to $GOPATH/src `git clone https://github.com/JLRgithub/PrivateDCi2b2.git` or...
- go get repository: `go get github.com/JLRgithub/PrivateDCi2b2`


## License

PDCi2b2 is licensed under a End User Software License Agreement ('EULA') for non-commercial use. If you want to have more information, please contact us.

## Contact
You can contact any of the developers for more information:

* [Jean Louis Raisaro](https://github.com/JLRgithub) (PHD student) - jean.raisaro@epfl.ch
* [Joao Andre Sa](https://github.com/JoaoAndreSa) (Software Engineer) - joao.gomesdesaesousa@epfl.ch
* [Jeffrey Klann](https://github.com/jklann) (Assistant Professor at HMS) - Jeff.Klann@mgh.harvard.edu