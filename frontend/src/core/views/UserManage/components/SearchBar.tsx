/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { Input, Flex, Button } from 'antd';
import { BsPersonFillAdd } from 'react-icons/bs';
import { useTranslation } from 'react-i18next';

interface SearchBarProps {
  username: string;
  corporation: string;
  onSearch: (type: 'username' | 'corporation', value: string) => void;
  onAddUser: () => void;
}

export const SearchBar: React.FC<SearchBarProps> = ({
  username,
  corporation,
  onSearch,
  onAddUser,
}) => {
  const { t } = useTranslation('core/userManage');

  return (
    <Flex className="h-[40px] w-full">
      <Flex className="w-full justify-between">
        <Flex className="w-full">
          <Flex className="w-auto flex items-center justify-start mr-5">
            <p className="text-md mr-2 mb-0">{t('index.userName')}：</p>
            <Input
              placeholder={t('index.search')}
              className="w-2/3"
              value={username}
              onChange={(e) => onSearch('username', e.target.value)}
            />
          </Flex>
          <Flex className="w-auto flex items-center justify-start">
            <p className="text-md mr-2 mb-0">{t('index.corporation')}：</p>
            <Input
              placeholder={t('index.search')}
              className="w-2/3"
              value={corporation}
              onChange={(e) => onSearch('corporation', e.target.value)}
            />
          </Flex>
        </Flex>
        <Flex className="w-full justify-end items-center">
          <Button
            type="primary"
            icon={<BsPersonFillAdd size={20} />}
            onClick={onAddUser}
            className="flex-grow-0 flex-shrink-0"
          >
            <span className="text-xs">{t('index.addUser')}</span>
          </Button>
        </Flex>
      </Flex>
    </Flex>
  );
};