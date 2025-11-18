import { createAlova } from 'alova';
import fetchAdapter from 'alova/fetch';
import { BASE_URL } from '@/utils/common';

export const alovaInstance = createAlova({
  baseURL: BASE_URL,
  requestAdapter: fetchAdapter(),
});

const getSuggestionUsers = (query: string) => {
  const params = {
    search: query,
  };
  return alovaInstance.Get('/api/v1/users/users/suggestions/', { params: params });
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

export { getSuggestionUsers, createShareURL, getShareSession, removeShareUser };
