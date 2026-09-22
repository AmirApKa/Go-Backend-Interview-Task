<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 flex items-center justify-center p-4" dir="rtl">
    <main class="w-full max-w-md bg-white rounded-2xl shadow-lg p-6 sm:p-8">
      <h1 class="text-2xl sm:text-3xl font-bold text-slate-800 text-center mb-2">
        تبدیل شماره کارت به شبا
      </h1>
      <p class="text-sm text-slate-500 text-center mb-6">
        شماره کارت ۱۶ رقمی خود را وارد کنید
      </p>

      <form @submit.prevent="handleSubmit" class="space-y-5" novalidate>
        <div>
          <label for="cardNumber" class="block text-sm font-medium text-slate-700 mb-2">
            شماره کارت:
          </label>
          <input
            id="cardNumber"
            v-model="cardNumber"
            type="text"
            inputmode="numeric"
            autocomplete="cc-number"
            maxlength="19"
            placeholder="۶۲۱۹ ۸۶۱۴ ۹۷۴۰ ۱۸۳۵"
            :disabled="loading"
            class="w-full px-4 py-3 border rounded-lg text-center text-lg tracking-widest font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 transition"
            :class="errorMessage ? 'border-red-400' : 'border-slate-300'"
          />
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full py-3 px-4 bg-blue-600 text-white font-semibold rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-slate-400 disabled:cursor-not-allowed transition-colors"
        >
          {{ loading ? 'در حال استعلام...' : 'استعلام شبا' }}
        </button>
      </form>

      <div v-if="ibanResult" class="mt-6 p-4 bg-green-50 border border-green-200 rounded-lg text-center">
        <p class="text-sm text-green-700 font-medium mb-1">شماره شبا:</p>
        <p class="text-lg font-bold text-green-800 tracking-wider font-mono" dir="ltr">{{ ibanResult }}</p>
      </div>

      <div v-if="errorMessage" class="mt-6 p-4 bg-red-50 border border-red-200 rounded-lg text-center">
        <p class="text-sm text-red-700">{{ errorMessage }}</p>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const config = useRuntimeConfig()

const cardNumber = ref('')
const ibanResult = ref('')
const errorMessage = ref('')
const loading = ref(false)

const normalizeCard = (value) => {
  const persian = '۰۱۲۳۴۵۶۷۸۹'
  const arabic = '٠١٢٣٤٥٦٧٨٩'
  return value
    .split('')
    .map((ch) => {
      const pi = persian.indexOf(ch)
      if (pi !== -1) return String(pi)
      const ai = arabic.indexOf(ch)
      if (ai !== -1) return String(ai)
      return ch
    })
    .join('')
    .replace(/[\s-]/g, '')
}

const handleSubmit = async () => {
  ibanResult.value = ''
  errorMessage.value = ''

  const normalized = normalizeCard(cardNumber.value)

  if (!normalized || normalized.length !== 16) {
    errorMessage.value = 'لطفاً یک شماره کارت ۱۶ رقمی وارد کنید.'
    return
  }

  if (!/^\d{16}$/.test(normalized)) {
    errorMessage.value = 'شماره کارت باید فقط شامل اعداد باشد.'
    return
  }

  loading.value = true

  try {
    const response = await $fetch('/api/v1/card-to-iban', {
      baseURL: config.public.apiBase,
      method: 'POST',
      body: { card_number: normalized }
    })

    if (response.success) {
      // فرمت‌بندی شبا همین‌جا (هر ۴ کاراکتر یه فاصله)
      ibanResult.value = response.iban.replace(/(.{4})/g, '$1 ').trim()
    } else {
      errorMessage.value = response.error || 'خطایی رخ داده است.'
    }
  } catch (error) {
    if (error?.data?.error) {
      errorMessage.value = error.data.error
    } else {
      errorMessage.value = 'ارتباط با سرور برقرار نشد. لطفاً دوباره تلاش کنید.'
    }
  } finally {
    loading.value = false
  }
}
</script>