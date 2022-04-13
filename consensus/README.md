# Consensus Module

The latest version of the consensus spec can be found [here](https://github.com/pokt-network/pocket-network-protocol/tree/main/consensus) and the changelog associated with this specific module can be found [here](./CHANGELOG.md).

## Demo

TODO: Insert video

## Env setup

Since

```
$ make docker_wipe
$ go mod vendor && go mod tidy
$ make mockgen
$ make protogen_local
$ make test_all
```

-     	$ make test_all
- ￼
-     	@OlshanskyI pulled P2P and the tests fine for me, make sure to run these commands first: $ go mod vendor && go mod tidy $ make mockgen $ make protogen_local $ make test_all

## Testing

### Localnet

First Shell:

```
$ make compose_and_watch
```

Second Shell:

```
$ make client_start
$ make client_connect
> ResetToGenesis
> PrintNodeState # Check committed height is 0
> TriggerNextView
> PrintNodeState # Check committed height is 1
> TriggerNextView
> PrintNodeState # Check committed height is 2
> TogglePacemakerMode # Check that it’s automatic now
> TriggerNextView # Let it rip!
```

## Navigating the code

```mermaid
graph TD;
    A-->B;
    A-->C;
    B-->D;
    C-->D;
```

```mermaid
erDiagram
    CUSTOMER ||--o{ ORDER : places
    ORDER ||--|{ LINE-ITEM : contains
    CUSTOMER }|..|{ DELIVERY-ADDRESS : uses
```

## Navigating the code

```mermaid
%% Example of sequence diagram
  sequenceDiagram
    Alice->>Bob: Hello Bob, how are you?
    alt is sick
    Bob->>Alice: Not so good :(
    else is well
    Bob->>Alice: Feeling fresh like a daisy
    end
    opt Extra response
    Bob->>Alice: Thanks for asking
    end
```

```mermaid
classDiagram
      Animal <|-- Duck
      Animal <|-- Fish
      Animal <|-- Zebra
      Animal : +int age
      Animal : +String gender
      Animal: +isMammal()
      Animal: +mate()
      class Duck{
          +String beakColor
          +swim()
          +quack()
      }
      class Fish{
          -int sizeInFeet
          -canEat()
      }
      class Zebra{
          +bool is_wild
          +run()
      }
```

< where do I start>

## Code Structure

< module structure>

```mermaid
graph LR

A & B--> C & D
style A fill:#f9f,stroke:#333,stroke-width:px
style B fill:#bbf,stroke:#f66,stroke-width:2px,color:#fff,stroke-dasharray: 5 5

subgraph beginning
A & B
end

subgraph ending
C & D
end

```

```mermaid
graph LR

A & B--> C & D
style A fill:#f9f,stroke:#333,stroke-width:px
style B fill:#bbf,stroke:#f66,stroke-width:2px,color:#fff,stroke-dasharray: 5 5

subgraph Pocket Node
P2P & Consensus & Utility & Persistence
end
```
