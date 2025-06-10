/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

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