import { Role } from "./role";

export interface User {
  userId: string | number;
  username: string;
  corporation?: string;
  email?: string;
  phone?: string;
  roleList?: Role[];
  role?: string; // For UI display
}