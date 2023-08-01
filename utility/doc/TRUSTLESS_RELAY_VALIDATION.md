# Validation of Trustless Relays

- [A. Client-side Relay Validation](#a-client-side-relay-validation)
- [B. Server-side Relay Validation](#b-server-side-relay-validation)

## A. Client-side Relay Validation

When an application requests to send a trustless relay, the CLI performs several checks on the relay before sending it to the specified servicer.
The following diagram lists all these checks with links to the corresponding code secion (or an issue if the check is not implemented yet).

```mermaid
---
title: Client-side Relay Validation
---
graph TD
    app_key{<b><a href='https://github.com/pokt-network/pocket/blob/bfc57eb89907932831b5730c68a09b4656c420c2/app/client/cli/servicer.go#L91'>Validate app key</a></b>}
    session{<b><a href='https://github.com/pokt-network/pocket/blob/bfc57eb89907932831b5730c68a09b4656c420c2/app/client/cli/servicer.go#L163'>Validate the Session</a></b>}
    servicer{<b><a href='https://github.com/pokt-network/pocket/blob/bfc57eb89907932831b5730c68a09b4656c420c2/app/client/cli/servicer.go#L133'>Validate the Servicer</a></b>}
    payload{<b><a href='https://github.com/pokt-network/pocket/blob/f41039b42ce628f73afe27b7f7b6111cca085cf0/app/client/cli/servicer.go#L191'>Deserialize Payload</a></b>}
    relay{<b><a href='https://github.com/pokt-network/pocket/issues/943'>Validate relay contents</a></b>}
    send[<b><a href='https://github.com/pokt-network/pocket/blob/f41039b42ce628f73afe27b7f7b6111cca085cf0/app/client/cli/servicer.go#L177'>Send Trustless Relay to the provided Servicer</a></b>]
    user_err[Return error to user]

    app_key-->|Failure| user_err
    session-->|Failure| user_err
    servicer-->|Failure| user_err
    payload-->|Failure| user_err
    relay-->|Failure| user_err

    app_key-->|Success| session
    session-->|Success| servicer
    servicer-->|Success| payload 
    payload-->|Success| relay
    relay-->|Success| send
```

## B. Server-side Relay Validation

Once a trustless relay has been received on the server side, i.e. by the servicer, several validations are performed on the relay. 
The following diagram outlines all these checks along with links to the corresponding section of the code (or to an issue if the check has not been implemented yet)

```mermaid
---
title: Server-side Relay Validation
---
graph TD
    deserialize{<b><a href='https://github.com/pokt-network/pocket/blob/f41039b42ce628f73afe27b7f7b6111cca085cf0/rpc/handlers.go#L85'>Deserialize Relay Payload</a></b>}
    meta{<b><a href='https://github.com/pokt-network/pocket/blob/f41039b42ce628f73afe27b7f7b6111cca085cf0/utility/servicer/module.go#L210'>Validate Relay Meta</a></b>}
    chain_support{<b><a href='https://github.com/pokt-network/pocket/blob/f41039b42ce628f73afe27b7f7b6111cca085cf0/utility/servicer/module.go#L221'>Validate chain support</a></b>}
    session{<b><a href='https://github.com/pokt-network/pocket/blob/f41039b42ce628f73afe27b7f7b6111cca085cf0/utility/servicer/module.go#L378'>Validate the Session</a></b>}
    height{<b><a href='https://github.com/pokt-network/pocket/blob/f41039b42ce628f73afe27b7f7b6111cca085cf0/utility/servicer/module.go#L496'>Validate Relay Height</a></b>}
    servicer{<b><a href='https://github.com/pokt-network/pocket/blob/f41039b42ce628f73afe27b7f7b6111cca085cf0/utility/servicer/module.go#L332'>Validate Servicer</a></b>}
    mine_relay{<b><a href='https://github.com/pokt-network/pocket/blob/f41039b42ce628f73afe27b7f7b6111cca085cf0/utility/servicer/module.go#L254'>Validate the app rate limit</a></b>}
    execute[<b><a href='https://github.com/pokt-network/pocket/blob/f41039b42ce628f73afe27b7f7b6111cca085cf0/utility/servicer/module.go#L191'>Execute the Relay</a></b>]
    client_err[Return error to client]

    deserialize-->|Failure| client_err
    meta-->|Failure| client_err
    chain_support-->|Failure| client_err
    session-->|Failure| client_err
    height-->|Failure| client_err
    servicer-->|Failure| client_err
    mine_relay-->|Failure| client_err
    
    deserialize-->|Success| meta
    meta-->|Success| chain_support
    chain_support-->|Success| session
    session-->|Success| height
    height-->|Success| servicer
    servicer-->|Success| mine_relay
    mine_relay-->|Success| execute    
```
<!-- GITHUB_WIKI: utility/trustless_relay -->
