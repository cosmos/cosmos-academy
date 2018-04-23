Testnets
========

brief blurbs then links to docs about it

Setup
------

-  \*Install **`**GNU Wget** <https://www.gnu.org/software/wget/>`__**:
   \*\*

**MacOS**

::

    brew install wget

**Linux**

::

    sudo apt-get install wget

Note: You can check other available options for downloading ``wget``
`here <https://www.gnu.org/software/wget/faq.html#download>`__.

**Get Source Code**

::

    go get github.com/cosmos/cosmos-sdk

Now we can fetch the correct versions of each dependency by running:

::

    ggmgit fetch --all
    git checkout develop
    make get_tools // run $ make update_tools if already installed
    make get_vendor_deps
    make install
    make install_examples

The latest cosmos-sdk should now be installed. Verify that everything is
OK by running:

::

    gaiad version

You should see:

::

    0.15.0-rc0-d613c2b9

And also:

::

    basecli version

You should see:

::

    0.15.0-rc0-d613c2b9

Genesis Setup
=============

Initiliaze Gaiad with the corresponding genesis files:

::

    gaiad init

Replace the genesis.json and config.toml files:

::

    rm $HOME/.gaiad/config/genesis.json $HOME/.gaiad/config/config.toml

    wget -O $HOME/.gaiad/config/genesis.json https://raw.githubusercontent.com/tendermint/testnets/master/gaia-3007/gaia/genesis.json

    wget -O $HOME/.gaiad/config/config.toml https://raw.githubusercontent.com/tendermint/testnets/master/gaia-3007/gaia/config.toml

Lastly change the ``moniker`` string in the\ ``config.toml``\ to
identify your node.

::

    # A custom human readable name for this node
    moniker = "<your_custom_name>"


    Starting Gaiad
    --------------

    Start the full node:

    ::

        gaiad start

    Check the everything is running smoothly:

    ::

        basecli status

    Generate keys
    -------------

    You'll need a private and public key pair (a.k.a. ``sk, pk``
    respectively) to be able to receive funds, send txs, bond tx, etc.

    To generate your keys:

    ::

        basecli keys add default // default is the name of the new generated private key

    Next, you will have to enter a passphrase for your ``default`` key. Keep
    the ``sk`` in a safe place.

    Now if you check your private keys you will see the ``default`` key
    among them:

    ::

        basecli keys show default

    You can see your other available keys by typing:

    ::

        basecli keys list

    *IMPORTANT: We strongly recommend to **NOT** use the same passphrase for
    your different keys. The Tendermint team and the Interchain Foundation
    will not be responsible for the lost of funds.*

    Getting Coins
    -------------

    Go to the faucet in http://atomexplorer.com/ and claim some coins for
    your testnet by typing the address of your key, as printed out above.

    Send tokens
    -----------

    ::

        basecli send --from=<your_address> --amount=1000fermion --sequence=1 --name=alice --to=5A35E4CC7B7DC0A5CB49CEA91763213A9AE92AD6

    The ``--amount`` flag defines the corresponding amount of the coin in
    the format ``--amount=<value|coin_name>``

    The ``--sequence`` flag corresponds to the sequence number to sign the
    tx.

    Now check the destination account and your own account to check the
    updated balances (by default the latest block):

    ::

        basecli account <destination_address>
        basecli account <your_address>

    You can also check your balance at a given block by using the
    ``--block`` flag:

    ::

        basecli account <your_address> --block=<block_height>

    Custom fee (coming soon)
    ~~~~~~~~~~~~~~~~~~~~~~~~

    You can also define a custom fee on the transaction by adding the
    ``--fee`` flag using the same format:

    ::

        basecli send --from=<your_address> --amount=1000fermion --fee=1fermion --sequence=1 --name=alice --to=5A35E4CC7B7DC0A5CB49CEA91763213A9AE92AD6

    Finally check your balance to see that your balance decreased:

    ::

        basecli account <your_validator_address_in_hex>

    Becoming a Validator
    --------------------

    Get your public key by typing:

    ::

        gaiad show_validator
        > 1624DE62201FF5974371065492BCD7E7E3212ABDD9145FAE53B6E062660F9433B97FC6B055

    The returned value is your validator address in hex. This can be used to
    create a new validator candidate:

    ::

        gaiacli declare-candidacy ...

    Staking
    ~~~~~~~

    Send the bonding staking transaction:

    ::

        basecli bond --stake=6steak --validator=<your_validator_address_in_hex> --sequence=0 --chain-id=<chain_name> --name=default

    Finally check your balance to see that your balance decreased:

    ::

        basecli account <your_validator_address_in_hex>

    Gaia Daemon
    ~~~~~~~~~~~

    Available commands

    ::

        // gaiad [command]
        help Help about any command
        init Initialize genesis files
        show_node_id Show this node's ID
        show_validator Show this node's validator info
        start Run the full node
        unsafe_reset_all Reset all blockchain data
        version Print the app version

    Basecoin light-client
    ~~~~~~~~~~~~~~~~~~~~~

    Available commands:

    ::

        // basecli [command]
        init Initialize light client
        status Query remote node for status
        block Get verified data for a the block at given height
        validatorset Get the full validator set at given height

        txs Search for all transactions that match the given tags
        tx Matches this txhash over all committed blocks

        account Query account balance
        send Create and sign a send tx
        transfer
        relay
        bond Bond to a validator
        unbond Unbond from a validator

        rest-server Start LCD (light-client daemon), a local REST server
        keys Add or view local private keys

        version Print the app version
        help Help about any command

    Add validator
    ~~~~~~~~~~~~~

    To get the information related to your validator node:

    ::

        gaiad show_validator

    Add a second validator candidate:

    ::

        basecli tx declare-candidacy --amount=10fermion --name=bob --pubkey=<pub_key data> --moniker=bobby

    Once that transaction is made, you should get an output like this one:

    ::

        Please enter passphrase for bob:
        {
        "check_tx": {
        "gas": 30
        },
        "deliver_tx": {},
        "hash": "2A2A61FFBA1D7A59138E0068C82CC830E5103799",
        "height": 4075
        }

    To check that the validator is active you can find it on the validator
    set list \*

    ::

        basecli validatorset <height>

    \*\ *Note: Remember that to be in the validator set you need to have
    more total power than the Xnd validator, where X is the assigned size
    for the validator set (by default *\ ``X = 100``\ *). *

    Delegating: Bonding and unbonding to a validator
    ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

    You can delegate (i.e. bind) **Atoms** to a validator to obtain a part
    of its fee revenue in exchange (the fee token in the Cosmos Hub are
    **Photons**). The command for delegating tokens is the same as staking
    just without the ``--stake`` flag:

    ::

        basecli bond --amount=10fermion --name=charlie --pubkey=<pub_key data>

    If for any reason the validator misbehaves or you just want to unbond a
    certain amount of the bonded tokens:

    ::

        basecli unbond --amount=5fermion --name=charlie --pubkey=<pub_key data>

    You should now see the unbonded tokens reflected in your balance:

    ::

        basecli account <your_address>

    Relaying
    ~~~~~~~~

    Relaying is key to enable interoperability in the Cosmos Ecosystem. It
    allows IBC packets of data to be sent from one chain to another. For a
    more deeper look into the Inter Blockchain Communication (IBC) protocol
    check this section.

    The command to relay packets is the following:

    ::

        basecli relay --from-chain-id=<origin_chain_name> --to-chain-id=<destination_chain_name> --from-chain-node=<host>:<port> --to-chain-node=<host>:<port> --name=<sk_to_sign_tx>
