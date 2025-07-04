import { createAlova } from 'alova';
import fetchAdapter from 'alova/fetch';
import { BASE_URL } from '@/utils/common';

export const alovaInstance = createAlova({
  baseURL: BASE_URL,
  requestAdapter: fetchAdapter(),
});

const GetUsersInfo = (query: string) => {
  const params = {
    action: 'suggestion',
    search: query,
  };
  return alovaInstance.Get('/api/v1/users/users/', { params: params });
};

const createShareURL = (data: any) => {
  return alovaInstance.Post(`/lion/api/share/`, data);
};

const getShareSession = (id: string, data: any) => {
  return alovaInstance.Post(`/lion/api/share/${id}/`, data);
};

const removeShareUser = (data: any) => {
  return alovaInstance.Post(`/lion/api/share/remove/`, data);
};

export { GetUsersInfo, createShareURL, getShareSession, removeShareUser };
