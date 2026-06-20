import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { tap } from 'rxjs';

export interface User {
  id: string;
  email: string;
  role: string;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private apiUrl = 'http://localhost:10000/api/v1/auth';
  
  currentUser = signal<User | null>(null);
  
  constructor(private http: HttpClient) {
    this.loadUserFromStorage();
  }
  
  login(email: string, password: string) {
    return this.http.post<{token: string, user: User}>(`${this.apiUrl}/login`, { email, password }).pipe(
      tap(res => {
        localStorage.setItem('eeip_token', res.token);
        localStorage.setItem('eeip_user', JSON.stringify(res.user));
        this.currentUser.set(res.user);
      })
    );
  }
  
  logout() {
    localStorage.removeItem('eeip_token');
    localStorage.removeItem('eeip_user');
    this.currentUser.set(null);
  }
  
  getToken(): string | null {
    return localStorage.getItem('eeip_token');
  }
  
  private loadUserFromStorage() {
    const userStr = localStorage.getItem('eeip_user');
    if (userStr) {
      try {
        const user = JSON.parse(userStr);
        this.currentUser.set(user);
      } catch (e) {
        this.logout();
      }
    }
  }
}
