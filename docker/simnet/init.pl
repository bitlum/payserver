#!/usr/bin/env perl

###############################################################################
#
# Simnet connector's dockers initialization script
#
# This script performs following operations:
# 1. sends half of generated funds from primary blockchain to secondary;
# 2. sends half of generated funds left from primary blockchain to lnd node;
# 4. establishes connection and opens channel between primary and secondary
#    lnd nodes;
#
# This script intended to be callend on docker machine host or after
# `docker-machine eval ...` / `docker-machine use ...`.
#
# This script allowed to be called multiple times but should be called once
# after `blocks-generator` initial blocks generation.
#
# Before calling this script you should wait until all containers are
# successfully run. Used `docker ps` to check containers statuses. All of
# them should be up for more than 1 minute.
#
# This script hasn't any external dependencies and requires only perl 5.22 or
# higher.
#
# Author: Vadim Chernov / dimuls@yandex.ru
#
###############################################################################

use v5.22;
use warnings;
use strict;

use JSON::PP qw/ decode_json /;

my $debug = 1;
my $account = 'default';
my $lightning_size = 15_000_000;

my @blockchains = qw/
	bitcoin
	bitcoin-cash
	dash
	ethereum
	litecoin
/;

my @lightning_blockchains = qw/
	bitcoin
/;

my @log_debug;

sub log_debug($) {
	push @log_debug, $_[0];
	say $_[0] if $debug;
}

$SIG{__DIE__} = sub {
	my ($error) = @_;
	unless ($debug) {
		say foreach @log_debug;
	}
	say $error;
};


sub trim($) {
	my ($s) = @_;
	$s =~ s/^\s+|\s+$//g;
	return $s;
}

sub trim_quotes($) {
	my ($s) = @_;
	$s =~ s/^"|"$//g;
	return $s
}

sub cmd_common($$$) {
	my ($blockchain, $role, $args) = @_;
	my $cmd = "docker exec -it $blockchain.simnet.$role $blockchain-cli ".trim(join(' ', split(/\s+/, $args)));
	log_debug "cmd: $cmd";
	my $res = `$cmd`;
	die "cmd: non zero exit code: $?. Output:\n$res"
		unless $? eq '0';
	return trim($res);
}

sub cmd_ethereum($$) {
	my ($role, $js) = @_;
	$js =~ s/'/\'/g;
	my $cmd = "docker exec -it ethereum.simnet.$role geth attach --exec '$js'";
	log_debug "cmd: $cmd";
	my $res = `$cmd`;
	die "cmd: non zero exit code: $?. Output:\n$res"
		unless $? eq '0';
	$res =~ s/\x1b\[[0-9;]*m//g;
	return trim($res);
}

sub cmd_lightning($$$) {
	my ($blockchain, $role, $args) = @_;
	my $cmd = "docker exec -it $blockchain-lightning.simnet.$role lncli ".trim(join(' ', split(/\s+/, $args)));
	log_debug "cmd: $cmd";
	my $res = `$cmd`;
	die "cmd: non zero exit code: $?. Output:\n$res"
		unless $? eq '0';
	return trim($res);
}

sub process_common($) {
	my ($blockchain) = @_;

	my $primary_funds = cmd_common($blockchain, 'primary', qq/ getbalance "" /);
	log_debug "$blockchain.simnet.primary \"\" account funds: $primary_funds";

	my $secondary_funds = cmd_common($blockchain, 'secondary', qq/ getbalance $account /);
	log_debug "$blockchain.simnet.secondary $account account funds: $secondary_funds";

	#
	# Sending funds to secondary default account
	#
	my $funds_diff = $primary_funds - $secondary_funds;
	if ($funds_diff <= 0) {
		log_debug "$blockchain.simnet.secondary has more funds when primary one: no need to deposit";
		return
	}

	my $secondary_addresses = decode_json(cmd_common($blockchain, 'secondary',
		qq/ getaddressesbyaccount "$account" /));

	my $secondary_address;
	if (@$secondary_addresses) {
		$secondary_address = $secondary_addresses->[0];
	} else {
		$secondary_address = cmd_common($blockchain, 'secondary',
			qq/ getnewaddress "$account" /);
	}

	log_debug "$blockchain.simnet.secondary deposit address: $secondary_address";

	my $deposit_funds = sprintf("%.3f", $funds_diff / 2);

	my $tx_id = cmd_common($blockchain, 'primary',
		qq/ sendtoaddress $secondary_address $deposit_funds /);

	log_debug "$blockchain.simnet.primary send $deposit_funds funds to secondary, tx_id: $tx_id";

	#
	# Sending funds to primary default account
	# 
	$primary_funds = cmd_common($blockchain, 'primary', qq/ getbalance "" /);
	log_debug "$blockchain.simnet.primary \"\" account funds: $primary_funds";

	my $primary_addresses = decode_json(cmd_common($blockchain, 'primary',
		qq/ getaddressesbyaccount "$account" /));

	my $primary_address;
	if (@$primary_addresses) {
		$primary_address = $primary_addresses->[0];
	} else {
		$primary_address = cmd_common($blockchain, 'primary',
			qq/ getnewaddress "$account" /);
	}

	log_debug "$blockchain.simnet.primary $account account deposit address: $primary_address";

	my $addlockconf_opt = "";
	$addlockconf_opt = "false" if $blockchain eq 'dash';

	$tx_id = cmd_common($blockchain, 'primary',
		qq/ sendmany "" "{\\"$primary_address\\":$primary_funds}" 1 $addlockconf_opt "" "[\\"$primary_address\\"]" /);

	log_debug "$blockchain.simnet.primary send $primary_funds from \"\" account to $account account, tx_id: $tx_id";
}

sub process_ethereum() {
	my $primary_funds = cmd_ethereum('primary', 'web3.fromWei(eth.getBalance(eth.coinbase), "ether")');
	log_debug "ethereum.simnet.primary funds: $primary_funds";

	my $secondary_funds = cmd_ethereum('secondary', 'web3.fromWei(eth.getBalance(eth.accounts[1]), "ether")');
	log_debug "ethereum.simnet.secondary funds: $secondary_funds";

	my $funds_diff = $primary_funds - $secondary_funds;
	if ($funds_diff <= 0) {
		log_debug "ethereum.simnet.secondary has more funds when primary one: no need to deposit";
		return;
	}

	my $secondary_address = trim_quotes cmd_ethereum('secondary', 'eth.accounts[1]');
	log_debug "ethereum.simnet.secondary deposit address: $secondary_address";

	my $deposit_funds = sprintf("%.3f", $funds_diff / 2);

	cmd_ethereum('primary', qq/ personal.unlockAccount(eth.coinbase, "") /);

	my $tx_id = trim_quotes cmd_ethereum('primary',
		qq/ eth.sendTransaction({from:eth.coinbase, to:"$secondary_address", value: web3.toWei($deposit_funds, "ether") }) /);
	
	log_debug "ethereum.simnet.primary send $deposit_funds funds to secondary, tx_id: $tx_id";
}

sub process_lightning($) {
	my ($blockchain) = @_;

	foreach my $role (qw/ primary secondary /) {
		my $blockchain_funds = cmd_common($blockchain, $role, qq/ getbalance /);
		if ($blockchain_funds == 0) {
			log_debug "$blockchain.simnet.$role has no funds, waiting...";
			while (1) {
				sleep 1;
				$blockchain_funds = cmd_common($blockchain, $role, qq/ getbalance /);
				last if $blockchain_funds > 0;
			}
		}
		log_debug "$blockchain.simnet.$role total funds: $blockchain_funds";

		my $lightning_funds = decode_json(cmd_lightning(
			$blockchain, $role, qq/ walletbalance /))->{total_balance};

		log_debug "$blockchain-lightning.simnet.$role funds: $lightning_funds";

		my $funds_diff = $blockchain_funds - $lightning_funds;
		if ($funds_diff <= 0) {
			log_debug "$blockchain-lightning.simnet.$role has more funds when $blockchain.simnet.$role: no need to deposit";
			next;
		}

		my $lightning_address = decode_json(cmd_lightning(
			$blockchain, $role, qq/ newaddress np2wkh /))->{address};

		log_debug "$blockchain-lightning.simnet.$role deposit address: $lightning_address";

		my $deposit_funds = sprintf("%.3f", $funds_diff / 2);

		my $tx_id = cmd_common($blockchain, $role,
				qq/ sendtoaddress $lightning_address $deposit_funds /);

		log_debug "$blockchain.simnet.primary send $deposit_funds funds to $blockchain-lightning.simnet.$role, tx_id: $tx_id";
	}

	foreach my $role (qw/ primary secondary /) {
		my $confirmed_funds = decode_json(cmd_lightning(
			$blockchain, $role, qq/ walletbalance /))->{confirmed_balance};

		log_debug "$blockchain-lightning.simnet.$role has no confirmed funds, waiting...";
		while (1) {
			sleep 1;
			$confirmed_funds = decode_json(cmd_lightning(
				$blockchain, $role, qq/ walletbalance /))->{confirmed_balance};
			last if $confirmed_funds > 0;
		}

		log_debug "$blockchain-lightning.simnet.$role confirmed funds: $confirmed_funds";
	}

	my $secondary_pubkey = decode_json(cmd_lightning($blockchain, 'secondary', 'getinfo'))->{identity_pubkey};
	log_debug "$blockchain-lightning.simnet.secondary identity pub key: $secondary_pubkey";

	my $primary_peers = decode_json(cmd_lightning($blockchain, 'primary', 'listpeers'))->{peers};
	log_debug "$blockchain-lightning.simnet.primary has ".scalar(@$primary_peers)." peers";

	my $connected = 0;
	foreach my $peer (@$primary_peers) {
		if ($peer->{pub_key} eq $secondary_pubkey) {
			$connected = 1;
			last;
		}
	}
	
	unless ($connected) {
		cmd_lightning($blockchain, 'primary',
			qq/ connect $secondary_pubkey\@bitcoin-lightning.simnet.secondary /);
		log_debug "$blockchain-lightning.simnet.primary has connected to secondary";
	}
	
	my $primary_channels = decode_json(cmd_lightning($blockchain, 'primary', 'listchannels'))->{channels};
	log_debug "$blockchain-lightning.simnet.primary has ".scalar(@$primary_channels)." channels";

	my $has_channel = 0;
	foreach my $channel (@$primary_channels) {
		if ($channel->{remote_pubkey} eq $secondary_pubkey) {
			$has_channel = 1;
			last;
		}
	}
	
	unless ($has_channel) {
		cmd_lightning($blockchain, 'primary',
			qq/ openchannel --block $secondary_pubkey $lightning_size /.($lightning_size/2));
		log_debug "$blockchain-lightning.simnet.primary has openned channel to secondary";
	}
}

sub main() {
	log_debug "simnet initialization stated";

	foreach my $blockchain (@blockchains) {
		log_debug "processing $blockchain blockchain";

		if ($blockchain eq 'ethereum') {
			process_ethereum();
		} else {
			process_common($blockchain);
		}

		log_debug "$blockchain blockchain processed";
	}

	foreach my $blockchain (@lightning_blockchains) {
		log_debug "processing $blockchain lightning";
		process_lightning($blockchain);
		log_debug "$blockchain lightning processed";
	}

	log_debug "simnet initialization completed";
}

main();

1;