# Source Generated with Decompyle++
# File: challenge_to_compile.pyc (Python 3.12)

import tkinter as tk
from tkinter import scrolledtext, messagebox, simpledialog, Checkbutton, BooleanVar, Toplevel
import platform
import hashlib
import time
import json
from threading import Thread
import math
import random
from Crypto.PublicKey import RSA
from Crypto.Util.number import bytes_to_long, long_to_bytes, isPrime
import os
import sys
from web3 import Web3
from eth_account import Account
from eth_account.signers.local import LocalAccount

def resource_path(relative_path):
    '''
    Get the absolute path to a resource, which works for both development
    and for a PyInstaller-bundled executable.
    '''
    base_path = sys._MEIPASS
    return os.path.join(base_path, relative_path)
# WARNING: Decompyle incomplete


class SmartContracts:
    rpc_url = ''
    private_key = ''
    
    def deploy_contract(contract_bytes, contract_abi):
        w3 = Web3(Web3.HTTPProvider(SmartContracts.rpc_url))
        if not w3.is_connected():
            raise ConnectionError(f'''[!] Failed to connect to Ethereum network at {SmartContracts.rpc_url}''')
        print(f'''[+] Connected to Sepolia network at {SmartContracts.rpc_url}''')
        print(f'''[+] Current block number: {w3.eth.block_number}''')
        if not SmartContracts.private_key:
            raise ValueError('Please add your private key.')
        account = Account.from_key(SmartContracts.private_key)
        w3.eth.default_account = account.address
        print(f'''[+] Using account: {account.address}''')
        balance_wei = w3.eth.get_balance(account.address)
        print(f'''[+] Account balance: {w3.from_wei(balance_wei, 'ether')} ETH''')
        if balance_wei == 0:
            print('[!] Warning: Account has 0 ETH. Deployment will likely fail. Get some testnet ETH from a faucet (e.g., sepoliafaucet.com)!')
        Contract = w3.eth.contract(abi = contract_abi, bytecode = contract_bytes)
        gas_estimate = Contract.constructor().estimate_gas()
        print(f'''[+] Estimated gas for deployment: {gas_estimate}''')
        gas_price = w3.eth.gas_price
        print(f'''[+] Current gas price: {w3.from_wei(gas_price, 'gwei')} Gwei''')
        transaction = Contract.constructor().build_transaction({
            'from': account.address,
            'nonce': w3.eth.get_transaction_count(account.address),
            'gas': gas_estimate + 200000,
            'gasPrice': gas_price })
        signed_txn = w3.eth.account.sign_transaction(transaction, private_key = SmartContracts.private_key)
        print('[+]  Deploying contract...')
        tx_hash = w3.eth.send_raw_transaction(signed_txn.raw_transaction)
        print(f'''[+] Deployment transaction sent. Hash: {tx_hash.hex()}''')
        print('[+] Waiting for transaction to be mined...')
        tx_receipt = w3.eth.wait_for_transaction_receipt(tx_hash, timeout = 300)
        print(f'''[+] Transaction receipt: {tx_receipt}''')
        if tx_receipt.status == 0:
            print('[!] Transaction failed (status 0). It was reverted.')
            return None
        contract_address = tx_receipt.contractAddress
        print(f'''[+] Contract deployed at address: {contract_address}''')
        deployed_contract = w3.eth.contract(address = contract_address, abi = contract_abi)
        return deployed_contract
    # WARNING: Decompyle incomplete



class LCGOracle:
    
    def __init__(self, multiplier, increment, modulus, initial_seed):
        self.multiplier = multiplier
        self.increment = increment
        self.modulus = modulus
        self.state = initial_seed
        self.contract_bytes = '6080604052348015600e575f5ffd5b506102e28061001c5f395ff3fe608060405234801561000f575f5ffd5b5060043610610029575f3560e01c8063115218341461002d575b5f5ffd5b6100476004803603810190610042919061010c565b61005d565b6040516100549190610192565b60405180910390f35b5f5f848061006e5761006d6101ab565b5b86868061007e5761007d6101ab565b5b8987090890505f5f8411610092575f610095565b60015b60ff16905081816100a69190610205565b858260016100b49190610246565b6100be9190610205565b6100c89190610279565b9250505095945050505050565b5f5ffd5b5f819050919050565b6100eb816100d9565b81146100f5575f5ffd5b50565b5f81359050610106816100e2565b92915050565b5f5f5f5f5f60a08688031215610125576101246100d5565b5b5f610132888289016100f8565b9550506020610143888289016100f8565b9450506040610154888289016100f8565b9350506060610165888289016100f8565b9250506080610176888289016100f8565b9150509295509295909350565b61018c816100d9565b82525050565b5f6020820190506101a55f830184610183565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601260045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f61020f826100d9565b915061021a836100d9565b9250828202610228816100d9565b9150828204841483151761023f5761023e6101d8565b5b5092915050565b5f610250826100d9565b915061025b836100d9565b9250828203905081811115610273576102726101d8565b5b92915050565b5f610283826100d9565b915061028e836100d9565b92508282019050808211156102a6576102a56101d8565b5b9291505056fea2646970667358221220c7e885c1633ad951a2d8168f80d36858af279d8b5fe2e19cf79eac15ecb9fdd364736f6c634300081e0033'
        self.contract_abi = [
            {
                'inputs': [
                    {
                        'internalType': 'uint256',
                        'name': 'LCG_MULTIPLIER',
                        'type': 'uint256' },
                    {
                        'internalType': 'uint256',
                        'name': 'LCG_INCREMENT',
                        'type': 'uint256' },
                    {
                        'internalType': 'uint256',
                        'name': 'LCG_MODULUS',
                        'type': 'uint256' },
                    {
                        'internalType': 'uint256',
                        'name': '_currentState',
                        'type': 'uint256' },
                    {
                        'internalType': 'uint256',
                        'name': '_counter',
                        'type': 'uint256' }],
                'name': 'nextVal',
                'outputs': [
                    {
                        'internalType': 'uint256',
                        'name': '',
                        'type': 'uint256' }],
                'stateMutability': 'pure',
                'type': 'function' }]
        self.deployed_contract = None

    
    def deploy_lcg_contract(self):
        self.deployed_contract = SmartContracts.deploy_contract(self.contract_bytes, self.contract_abi)

    
    def get_next(self, counter):
        print(f'''\n[+] Calling nextVal() with _currentState={self.state}''')
        self.state = self.deployed_contract.functions.nextVal(self.multiplier, self.increment, self.modulus, self.state, counter).call()
        print(f'''  _counter = {counter}: Result = {self.state}''')
        return self.state



class TripleXOROracle:
    
    def __init__(self):
        self.contract_bytes = '61030f61004d600b8282823980515f1a6073146041577f4e487b71000000000000000000000000000000000000000000000000000000005f525f60045260245ffd5b305f52607381538281f3fe7300000000000000000000000000000000000000003014608060405260043610610034575f3560e01c80636230075614610038575b5f5ffd5b610052600480360381019061004d919061023c565b610068565b60405161005f91906102c0565b60405180910390f35b5f5f845f1b90505f845f1b90505f61007f85610092565b9050818382181893505050509392505050565b5f5f8290506020815111156100ae5780515f525f5191506100b6565b602081015191505b50919050565b5f604051905090565b5f5ffd5b5f5ffd5b5f819050919050565b6100df816100cd565b81146100e9575f5ffd5b50565b5f813590506100fa816100d6565b92915050565b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b61014e82610108565b810181811067ffffffffffffffff8211171561016d5761016c610118565b5b80604052505050565b5f61017f6100bc565b905061018b8282610145565b919050565b5f67ffffffffffffffff8211156101aa576101a9610118565b5b6101b382610108565b9050602081019050919050565b828183375f83830152505050565b5f6101e06101db84610190565b610176565b9050828152602081018484840111156101fc576101fb610104565b5b6102078482856101c0565b509392505050565b5f82601f83011261022357610222610100565b5b81356102338482602086016101ce565b91505092915050565b5f5f5f60608486031215610253576102526100c5565b5b5f610260868287016100ec565b9350506020610271868287016100ec565b925050604084013567ffffffffffffffff811115610292576102916100c9565b5b61029e8682870161020f565b9150509250925092565b5f819050919050565b6102ba816102a8565b82525050565b5f6020820190506102d35f8301846102b1565b9291505056fea26469706673582212203fc7e6cc4bf6a86689f458c2d70c565e7c776de95b401008e58ca499ace9ecb864736f6c634300081e0033'
        self.contract_abi = [
            {
                'inputs': [
                    {
                        'internalType': 'uint256',
                        'name': '_primeFromLcg',
                        'type': 'uint256' },
                    {
                        'internalType': 'uint256',
                        'name': '_conversationTime',
                        'type': 'uint256' },
                    {
                        'internalType': 'string',
                        'name': '_plaintext',
                        'type': 'string' }],
                'name': 'encrypt',
                'outputs': [
                    {
                        'internalType': 'bytes32',
                        'name': '',
                        'type': 'bytes32' }],
                'stateMutability': 'pure',
                'type': 'function' }]
        self.deployed_contract = None

    
    def deploy_triple_xor_contract(self):
        self.deployed_contract = SmartContracts.deploy_contract(self.contract_bytes, self.contract_abi)

    
    def encrypt(self, prime_from_lcg, conversation_time, plaintext_bytes):
        print(f'''\n[+] Calling encrypt() with prime_from_lcg={prime_from_lcg}, time={conversation_time}, plaintext={plaintext_bytes}''')
        ciphertext = self.deployed_contract.functions.encrypt(prime_from_lcg, conversation_time, plaintext_bytes).call()
        print(f'''  _ciphertext = {ciphertext.hex()}''')
        return ciphertext



class ChatLogic:
    
    def __init__(self):
        self.lcg_oracle = None
        self.xor_oracle = None
        self.rsa_key = None
        self.seed_hash = None
        self.super_safe_mode = False
        self.message_count = 0
        self.conversation_start_time = 0
        self.chat_history = []
        self._initialize_crypto_backend()

    
    def _get_system_artifact_hash(self):
        artifact = platform.node().encode('utf-8')
        hash_val = hashlib.sha256(artifact).digest()
        seed_hash = int.from_bytes(hash_val, 'little')
        print(f'''[SETUP]  - Generated Seed {seed_hash}...''')
        return seed_hash

    
    def _generate_primes_from_hash(self, seed_hash):
        primes = []
        current_hash_byte_length = (seed_hash.bit_length() + 7) // 8
        current_hash = seed_hash.to_bytes(current_hash_byte_length, 'little')
        print('[SETUP] Generating LCG parameters from system artifact...')
        iteration_limit = 10000
        iterations = 0
        if len(primes) < 3 and iterations < iteration_limit:
            current_hash = hashlib.sha256(current_hash).digest()
            candidate = int.from_bytes(current_hash, 'little')
            iterations += 1
            if candidate.bit_length() == 256 and isPrime(candidate):
                primes.append(candidate)
                print(f'''[SETUP]  - Found parameter {len(primes)}: {str(candidate)[:20]}...''')
            if len(primes) < 3 and iterations < iteration_limit:
                continue
        if len(primes) < 3:
            error_msg = '[!] Error: Could not find 3 primes within iteration limit.'
            print('Current Primes: ', primes)
            print(error_msg)
            exit()
        return (primes[0], primes[1], primes[2])

    
    def _initialize_crypto_backend(self):
        self.seed_hash = self._get_system_artifact_hash()
        (m, c, n) = self._generate_primes_from_hash(self.seed_hash)
        self.lcg_oracle = LCGOracle(m, c, n, self.seed_hash)
        self.lcg_oracle.deploy_lcg_contract()
        print('[SETUP] LCG Oracle is on-chain...')
        self.xor_oracle = TripleXOROracle()
        self.xor_oracle.deploy_triple_xor_contract()
        print('[SETUP] Triple XOR Oracle is on-chain...')
        print('[SETUP] Crypto backend initialized...')

    
    def generate_rsa_key_from_lcg(self):
        print('[RSA] Generating RSA key from on-chain LCG primes...')
        lcg_for_rsa = LCGOracle(self.lcg_oracle.multiplier, self.lcg_oracle.increment, self.lcg_oracle.modulus, self.seed_hash)
        lcg_for_rsa.deploy_lcg_contract()
        primes_arr = []
        rsa_msg_count = 0
        iteration_limit = 10000
        iterations = 0
        if len(primes_arr) < 8 and iterations < iteration_limit:
            candidate = lcg_for_rsa.get_next(rsa_msg_count)
            rsa_msg_count += 1
            iterations += 1
            if candidate.bit_length() == 256 and isPrime(candidate):
                primes_arr.append(candidate)
                print(f'''[RSA]  - Found 256-bit prime #{len(primes_arr)}''')
            if len(primes_arr) < 8 and iterations < iteration_limit:
                continue
        print('Primes Array: ', primes_arr)
        if len(primes_arr) < 8:
            error_msg = '[RSA] Error: Could not find 8 primes within iteration limit.'
            print('Current Primes: ', primes_arr)
            print(error_msg)
            return error_msg
        n = None
        for p_val in primes_arr:
            n *= p_val
        phi = 1
        for p_val in primes_arr:
            phi *= p_val - 1
        e = 65537
        if math.gcd(e, phi) != 1:
            error_msg = '[RSA] Error: Public exponent e is not coprime with phi(n). Cannot generate key.'
            print(error_msg)
            return error_msg
        self.rsa_key = None.construct((n, e))
    # WARNING: Decompyle incomplete

    
    def process_message(self, plaintext):
        if self.conversation_start_time == 0:
            self.conversation_start_time = time.time()
        conversation_time = int(time.time() - self.conversation_start_time)
        if self.super_safe_mode and self.rsa_key:
            plaintext_bytes = plaintext.encode('utf-8')
            plaintext_enc = bytes_to_long(plaintext_bytes)
            _enc = pow(plaintext_enc, self.rsa_key.e, self.rsa_key.n)
            ciphertext = _enc.to_bytes(self.rsa_key.n.bit_length(), 'little').rstrip(b'\x00')
            encryption_mode = 'RSA'
            plaintext = '[ENCRYPTED]'
        else:
            prime_from_lcg = self.lcg_oracle.get_next(self.message_count)
            ciphertext = self.xor_oracle.encrypt(prime_from_lcg, conversation_time, plaintext)
            encryption_mode = 'LCG-XOR'
        log_entry = {
            'conversation_time': conversation_time,
            'mode': encryption_mode,
            'plaintext': plaintext,
            'ciphertext': ciphertext.hex() }
        self.chat_history.append(log_entry)
        self.save_chat_log()
        return (f'''[{conversation_time}s] {plaintext}''', f'''[{conversation_time}s] {ciphertext.hex()}''')

    
    def save_chat_log(self):
        pass
    # WARNING: Decompyle incomplete



class ChatApp(tk.Tk):
    pass
# WARNING: Decompyle incomplete

if __name__ == '__main__':
    app = ChatApp()
    app.mainloop()
    return None
