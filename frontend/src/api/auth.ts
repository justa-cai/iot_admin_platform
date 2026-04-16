import client from './client'
import type { LoginRequest, LoginResponse, RegisterRequest, User } from '@/types'

export function login(data: LoginRequest): Promise<LoginResponse> {
  return client.post('/auth/login', data)
}

export function register(data: RegisterRequest): Promise<User> {
  return client.post('/auth/register', data)
}

export function getProfile(): Promise<User> {
  return client.get('/auth/profile')
}
