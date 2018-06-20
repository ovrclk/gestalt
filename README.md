<p align="center">
  <img src="https://i.imgur.com/JTsSMat.png" />
</p>

# Gestalt

Gestalt is an integration test framework built for testing
CLI-based workflows.


###

component-list: component :: component-list?

component: 

  - name: name
    type: type
    args: [ args1 , ...argsN  ]
    run:  [ child1, ...childN ]

  or

  - type: [ arg1, ...argN ]

  or

  - type: [ child1, ...childN ]

  or

  - type: name
    args: [ args1 , ...argsN  ]
    run:  [ child1, ...childN ]

  or

  - name

type: defaults to 'group'

todo: add retry, ignore to all components
