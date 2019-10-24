#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import requests

DDNS_SERVICE = "dyndns.example.org"
SECRET = "abc1234"
SERVER_NAME = "home"


def update_dns():
    response = requests.get(
        'https://{}/update?secret={}&domain={}'.format(DDNS_SERVICE,
                                                       SECRET,
                                                       SERVER_NAME))

    return response


def main():
    dns = update_dns().json()

    if dns['Success']:
        if dns['Message'] != 'Record exist already':
            print(
                'Update DNS for {}.{} with ip {}'.format(SERVER_NAME,
                                                         DDNS_SERVICE,
                                                         dns['Address']))
    else:
        print('DNS update for {}.{} faild!'.format(SERVER_NAME, DDNS_SERVICE))


if __name__ == '__main__':
    main()
