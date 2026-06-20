import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthService } from '../auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './login.html',
  styleUrls: ['./login.css']
})
export class Login {
  isLoginMode = true;
  email = '';
  password = '';
  accountName = '';
  imapHost = '';
  imapPort = 993;
  imapUser = '';
  imapPassword = '';
  
  errorMessage = '';
  isLoading = false;

  constructor(private authService: AuthService) {}

  toggleMode() {
    this.isLoginMode = !this.isLoginMode;
    this.errorMessage = '';
  }

  submit() {
    if (this.isLoginMode) {
      this.login();
    } else {
      this.register();
    }
  }

  login() {
    if (!this.email || !this.password) {
      this.errorMessage = 'Please enter both email and password.';
      return;
    }
    this.isLoading = true;
    this.errorMessage = '';
    
    this.authService.login(this.email, this.password).subscribe({
      next: () => {
        this.isLoading = false;
        // The parent component (AppComponent) will react to authService.currentUser
      },
      error: (err) => {
        this.isLoading = false;
        this.errorMessage = err.error?.message || err.error?.error || 'Login failed. Please check your credentials.';
      }
    });
  }

  register() {
    if (!this.email || !this.password || !this.accountName || !this.imapHost || !this.imapUser || !this.imapPassword) {
      this.errorMessage = 'Please fill out all required fields.';
      return;
    }
    this.isLoading = true;
    this.errorMessage = '';
    
    const payload = {
      email: this.email,
      password: this.password,
      account_name: this.accountName,
      imap_host: this.imapHost,
      imap_port: Number(this.imapPort),
      imap_user: this.imapUser,
      imap_password: this.imapPassword
    };

    this.authService.register(payload).subscribe({
      next: () => {
        // Automatically login
        this.authService.login(this.email, this.password).subscribe({
          next: () => {
            this.isLoading = false;
          },
          error: (err) => {
            this.isLoading = false;
            this.errorMessage = 'Registration successful but auto-login failed.';
          }
        });
      },
      error: (err) => {
        this.isLoading = false;
        this.errorMessage = err.error?.message || err.error?.error || 'Registration failed.';
      }
    });
  }
}
