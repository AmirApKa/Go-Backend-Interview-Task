<template>
  <div class="container">
    <h2>تبدیل شماره کارت به شبا</h2>
    
    <form @submit.prevent="handleSubmit">
      <div class="form-group">
        <label>شماره کارت ۱۶ رقمی:</label>
        <input 
          v-model="cardNumber" 
          type="text" 
          maxlength="16" 
          placeholder="مثال: ۶۲۱۹۸۶۱۴۹۷۴۰۱۸۳۵"
          :disabled="loading"
        />
      </div>

      <button type="submit" :disabled="loading">
        {{ loading ? 'در حال استعلام...' : 'استعلام شبا' }}
      </button>
    </form>

    <div v-if="ibanResult" class="result success">
      <p><strong>شماره شبا:</strong> {{ ibanResult }}</p>
    </div>

    <div v-if="errorMessage" class="result error">
      <p>{{ errorMessage }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'

const cardNumber = ref('')
const ibanResult = ref('')
const errorMessage = ref('')
const loading = ref(false)

const handleSubmit = async () => {
  ibanResult.value = ''
  errorMessage.value = ''

  if (!cardNumber.value || cardNumber.value.length !== 16) {
    errorMessage.value = 'لطفاً یک شماره کارت ۱۶ رقمی وارد کنید.'
    return
  }

  loading.value = true

  try {
    const response = await axios.post('http://localhost:8080/api/v1/card-to-iban', {
      card_number: cardNumber.value
    })

    if (response.data.success) {
      ibanResult.value = response.data.iban
    } else {
      errorMessage.value = response.data.error || 'خطایی رخ داده است.'
    }
  } catch (error) {
    if (error.response && error.response.data) {
      errorMessage.value = error.response.data.error || 'خطا در برقراری ارتباط با سرور'
    } else {
      errorMessage.value = 'ارتباط با سرور برقرار نشد.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.container {
  max-width: 400px;
  margin: 50px auto;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  direction: rtl;
  font-family: sans-serif;
}
.form-group {
  margin-bottom: 15px;
}
input {
  width: 100%;
  padding: 10px;
  margin-top: 5px;
  box-sizing: border-box;
  font-size: 16px;
  text-align: center;
}
button {
  width: 100%;
  padding: 10px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
}
button:disabled {
  background-color: #ccc;
}
.result {
  margin-top: 20px;
  padding: 10px;
  border-radius: 4px;
  text-align: center;
}
.success {
    background-color: #d4edda;
    color: #155724;
  }
.error {
  background-color: #f8d7da;
  color: #721c24;
}
</style>