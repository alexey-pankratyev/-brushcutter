# Brushcutter
___
 Scripts helpers for kubernetes
 ___
 #### Rights_namespaces
> _This script to help creat rolebindings in multi contexts_

_Befor launch you need to create config.yml file with maps of Rolebindings_
 
example: 
```
rolebindings:
  cluster-context1: 
    - namespace1
    - namespace2
  cluster-context2:
    - namespace1
    - namespace2
    - namespace3
```
 
