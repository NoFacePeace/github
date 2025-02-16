package array

// https://leetcode.cn/problems/replace-elements-with-greatest-element-on-right-side/

func replaceElements(arr []int) []int {
    n := len(arr)
    ans := make([]int, n)
    for i := n - 1; i >= 0; i-- {
        if i == n - 1 {
            ans[i] = -1
            continue
        }
        if i == n - 2 {
            ans[i] = arr[i+1]
            continue
        }
        if arr[i+1] > ans[i+1] {
            ans[i] = arr[i+1]
        } else {
            ans[i] = ans[i+1]
        }
    }
    return ans
}