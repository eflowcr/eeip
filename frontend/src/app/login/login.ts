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

  onEmailChange(newEmail: string) {
    this.email = newEmail;
    if (!this.isLoginMode) {
      this.imapUser = newEmail;
      this.imapPort = 993;
      
      const parts = newEmail.split('@');
      if (parts.length === 2 && parts[1]) {
        const domain = parts[1].toLowerCase();
        
        this.accountName = newEmail;

        if (domain === 'gmail.com') {
          this.imapHost = 'imap.gmail.com';
        } else if (domain === 'outlook.com' || domain === 'hotmail.com') {
          this.imapHost = 'outlook.office365.com';
        } else if (domain === 'yahoo.com') {
          this.imapHost = 'imap.mail.yahoo.com';
        } else {
          this.imapHost = `mail.${domain}`;
        }
      } else {
        this.accountName = newEmail;
        this.imapHost = '';
      }
    }
  }

  onPasswordChange(newPassword: string) {
    this.password = newPassword;
    if (!this.isLoginMode) {
      this.imapPassword = newPassword;
    }
  }

  toggleMode() {
    this.isLoginMode = !this.isLoginMode;
    this.errorMessage = '';
    if (!this.isLoginMode && this.email) {
      this.onEmailChange(this.email);
    }
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
